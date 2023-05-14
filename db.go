package p2p_database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	ipfs_datastore "github.com/ipfs/go-datastore/sync"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"strings"
	"sync"
	"time"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"

	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/dual"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	crdt "github.com/ipfs/go-ds-crdt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultPort = 3500

	DefaultDatabaseEventsBufferSize = 128
	RebroadcastingInterval          = 1 * time.Second

	NetSubscriptionTopicPrefix  = "crdt_net_"
	NetSubscriptionPublishValue = "ping"
)

var (
	ErrEmptyKey                     = errors.New("empty key")
	ErrKeyNotFound                  = errors.New("key not found")
	ErrEthereumWalletNotRegistered  = errors.New("ethereum address not registered")
	ErrIncorrectSubscriptionHandler = errors.New("incorrect subscription handler")
)

var (
	onceInitHostP2P      = sync.Once{}
	globalHost           host.Host
	globalDHT            *dual.DHT
	globalBootstrapNodes []peer.AddrInfo
)

type PubSubHandler func(Event)

type TopicSubscription struct {
	subscription *pubsub.Subscription
	topic        *pubsub.Topic
	handler      PubSubHandler
}

type Event struct {
	ID         string
	FromPeerId string
	Message    interface{}
}

type DB struct {
	Name             string
	selfID           peer.ID
	host             host.Host
	crdt             *crdt.Datastore
	ethSmartContract *EthSmartContract

	ds                 datastore.Batching
	pubSub             *pubsub.PubSub
	joinedTopics       map[string]*pubsub.Topic
	topicSubscriptions map[string]*TopicSubscription
	handleGroup        *errgroup.Group
	lock               sync.RWMutex

	netTopic        *pubsub.Topic
	netSubscription *pubsub.Subscription

	disconnectOnce sync.Once
	logger         *logging.ZapEventLogger
}

func Connect(
	ctx context.Context,
	config Config,
	logger *logging.ZapEventLogger,
	opts ...dht.Option,
) (*DB, error) {
	crypto.MinRsaKeyBits = 1024

	grp := &errgroup.Group{}
	grp.SetLimit(DefaultDatabaseEventsBufferSize)

	ethSmartContract, err := NewEthSmartContract(config, logger)
	if err != nil {
		return nil, errors.Wrap(err, "create ethereum smart contract")
	}

	port := EnvConfig.PeerListenPort
	if port == 0 {
		port = DefaultPort
	}
	h, kdht, err := makeHost(ctx, ethSmartContract, config.WalletPrivateKey, port)
	if err != nil {
		return nil, errors.Wrap(err, "make lib p2p host")
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, errors.Wrap(err, "create pubsub")
	}

	pubsubBC, err := crdt.NewPubSubBroadcaster(ctx, ps, "crdt_"+config.DatabaseName)
	if err != nil {
		return nil, errors.Wrap(err, "init pub sub crdt broadcaster")
	}

	ds := ipfs_datastore.MutexWrap(datastore.NewMapDatastore())
	ipfs, err := ipfslite.New(ctx, ds, nil, h, kdht, nil)
	if err != nil {
		return nil, errors.Wrap(err, "init ipfs")
	}
	ipfs.Bootstrap(globalBootstrapNodes)

	for i, bootstrapNode := range globalBootstrapNodes {
		logger.Infof("Bootstrap node %d - %s - [%s]\n\n", i, bootstrapNode.String(), bootstrapNode.Addrs[0].String())
		h.ConnManager().TagPeer(bootstrapNode.ID, "keep", 100)
	}

	valid, err := ethSmartContract.ValidatePeer(h.ID())
	if err != nil {
		return nil, errors.Wrap(err, "try validate current peer id in smart contract")
	}
	if !valid {
		return nil, ErrEthereumWalletNotRegistered
	}

	logging.SetLogLevel("globaldb", "debug")

	crtdOpts := crdt.DefaultOptions()
	crtdOpts.Logger = logging.Logger("globaldb")
	crtdOpts.RebroadcastInterval = RebroadcastingInterval
	crtdOpts.PutHook = func(k datastore.Key, v []byte) {
		fmt.Printf("Added: [%s] -> %s\n", k, string(v))
	}
	crtdOpts.DeleteHook = func(k datastore.Key) {
		fmt.Printf("Removed: [%s]\n", k)
	}

	datastoreCrdt, err := crdt.New(ds, datastore.NewKey("crdt_"+config.DatabaseName), ipfs, pubsubBC, crtdOpts)
	if err != nil {
		return nil, errors.Wrap(err, "init crdt")
	}

	err = datastoreCrdt.Sync(ctx, datastore.NewKey("/"))
	if err != nil {
		return nil, errors.Wrap(err, "crdt sync datastore")
	}

	netTopic, err := ps.Join(NetSubscriptionTopicPrefix + config.DatabaseName)
	if err != nil {
		return nil, errors.Wrap(err, "create net topic")
	}

	netSubscription, err := netTopic.Subscribe()
	if err != nil {
		return nil, errors.Wrap(err, "subscribe to net topic")
	}

	db := &DB{
		Name:   config.DatabaseName,
		host:   h,
		selfID: h.ID(),
		logger: logger,

		ds:               datastoreCrdt,
		ethSmartContract: ethSmartContract,
		crdt:             datastoreCrdt,

		pubSub:             ps,
		topicSubscriptions: map[string]*TopicSubscription{},
		joinedTopics:       map[string]*pubsub.Topic{},
		handleGroup:        grp,
		lock:               sync.RWMutex{},

		netTopic:        netTopic,
		netSubscription: netSubscription,

		disconnectOnce: sync.Once{},
	}
	db.refreshPeers(ctx)

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		db.Disconnect(ctx)
	}()

	return db, nil
}

func (db *DB) String() string {
	return db.Name
}

func (db *DB) List(ctx context.Context) ([]string, error) {
	r, err := db.crdt.Query(ctx, query.Query{KeysOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "crdt list query")
	}

	var keys []string
	for k := range r.Next() {
		keys = append(keys, k.Key)
	}

	return keys, nil
}

func (db *DB) Set(ctx context.Context, key, value string) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	err := db.crdt.Put(ctx, datastore.NewKey(key), []byte(value))
	if err != nil {
		return errors.Wrap(err, "crdt put")
	}

	return nil
}

func (db *DB) Get(ctx context.Context, key string) (string, error) {
	if len(key) == 0 {
		return "", ErrEmptyKey
	}

	val, err := db.crdt.Get(ctx, datastore.NewKey(key))
	switch {
	case err != nil && strings.Contains(err.Error(), "key not found"):
		return "", ErrKeyNotFound
	case err != nil:
		return "", errors.Wrap(err, "crdt get")
	}

	return string(val), nil
}

func (db *DB) Remove(ctx context.Context, key string) error {
	err := db.crdt.Delete(ctx, datastore.NewKey(key))
	switch {
	case err != nil && strings.Contains(err.Error(), "key not found"):
		return ErrKeyNotFound
	case err != nil:
		return errors.Wrap(err, "crdt delete")
	}
	return nil
}

func (db *DB) Subscribe(ctx context.Context, topic string, handler PubSubHandler, opts ...pubsub.TopicOpt) error {
	t, err := db.joinTopic(topic, opts...)
	if err != nil {
		return err
	}

	s, err := t.Subscribe()
	if err != nil {
		return errors.Wrap(err, "pub sub subscribe topic")
	}

	if handler == nil {
		return ErrIncorrectSubscriptionHandler
	}

	db.lock.Lock()
	db.topicSubscriptions[topic] = &TopicSubscription{
		subscription: s,
		topic:        t,
		handler:      handler,
	}
	db.lock.Unlock()

	db.handleGroup.Go(func() error {
		err = db.listenEvents(ctx, db.topicSubscriptions[topic])
		if err != nil {
			db.logger.Errorf("pub sub listen events topic %s err %s", topic, err)
		}
		return err
	})

	return nil
}

func (db *DB) Publish(ctx context.Context, topic string, value interface{}, opts ...pubsub.PubOpt) (Event, error) {
	t, err := db.joinTopic(topic)
	if err != nil {
		return Event{}, err
	}

	event := Event{
		ID:         uuid.New().String(),
		Message:    value,
		FromPeerId: db.host.ID().String(),
	}
	marshaled, err := json.Marshal(event)
	if err != nil {
		return Event{}, errors.Wrap(err, "try marshal message")
	}

	err = t.Publish(ctx, marshaled, opts...)
	if err != nil {
		return Event{}, errors.Wrap(err, "pub sub publish message")
	}

	return event, nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	var err error
	db.disconnectOnce.Do(func() {
		err = db.disconnect(ctx)
	})
	return err
}

func (db *DB) disconnect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return db.handleGroup.Wait()
	})
	g.Go(func() error {
		err := db.ds.Close()
		if err != nil {
			return errors.Wrap(err, "datastore close")
		}
		return nil
	})
	g.Go(func() error {
		err := db.crdt.Close()
		if err != nil {
			return errors.Wrap(err, "crdt close")
		}
		return nil
	})
	g.Go(func() error {
		err := db.host.Close()
		if err != nil {
			return errors.Wrap(err, "host close")
		}
		return nil
	})
	g.Go(func() error {
		err := globalDHT.Close()
		if err != nil {
			return errors.Wrap(err, "globalDHT close")
		}
		return nil
	})
	g.Go(func() error {
		db.lock.RLock()
		for _, s := range db.topicSubscriptions {
			s.subscription.Cancel()
			err := s.topic.Close()
			if err != nil {
				db.logger.Errorf("try close db topic %s current peer id %s", s.topic, db.host.ID())
			}
		}
		db.lock.RUnlock()
		return nil
	})

	return g.Wait()
}

func (db *DB) GetHost() host.Host {
	return db.host
}

func (db *DB) joinTopic(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	ts, ok := db.topicSubscriptions[topic]
	//already joined
	if ok {
		return ts.topic, nil
	}

	if t, ok := db.joinedTopics[topic]; ok {
		return t, nil
	}

	t, err := db.pubSub.Join(topic, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "pub sub join topic")
	}
	db.joinedTopics[topic] = t

	return t, nil
}

func (db *DB) listenEvents(ctx context.Context, topicSub *TopicSubscription) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := topicSub.subscription.Next(ctx)
			if err != nil {
				db.logger.Errorf("try get next pub sub message error: %s", err)

				continue
			}

			//skip self messages
			if msg.ReceivedFrom == db.selfID {
				if msg.Message != nil {
					db.logger.Debugf("fetched message from self publish %s", string(msg.Message.Data))
				}
				continue
			}

			event := Event{}
			err = json.Unmarshal(msg.Data, &event)
			if err != nil {
				db.logger.Errorf("try unmarshal pub sub message from %s error %s, data: %s", msg.ReceivedFrom, err, string(msg.Data))
			}

			topicSub.handler(event)
		}
	}
}

func (db *DB) refreshPeers(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := db.netSubscription.Next(ctx)
				if err != nil {
					db.logger.Errorf("try net subscription read next message: %s", err)
					continue
				}

				if msg.ReceivedFrom == db.host.ID() {
					continue
				}

				db.host.ConnManager().TagPeer(msg.ReceivedFrom, "keep", 100)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := db.netTopic.Publish(ctx, []byte(NetSubscriptionPublishValue))
				if err != nil {
					db.logger.Errorf("try publish message to net ps topic: %s", err)
				}
				time.Sleep(20 * time.Second)
			}
		}
	}()
}

func makeHost(ctx context.Context, ethSmartContract *EthSmartContract, ethPrivateKey string, port int) (host.Host, *dual.DHT, error) {
	prvKey, err := eth_crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "hex to ecdsa eth private key")
	}

	privKeyBytes := eth_crypto.FromECDSA(prvKey)
	priv, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "UnmarshalSecp256k1PrivateKey from eth private key")
	}

	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		return nil, nil, errors.Wrap(err, "create multi addr")
	}

	var errSetupLibP2P error
	onceInitHostP2P.Do(func() {
		opts := ipfslite.Libp2pOptionsExtra
		opts = append(
			opts,
			libp2p.ConnectionGater(
				NewEthConnectionGater(ethSmartContract, *logging.Logger("eth-connection-gater")),
			),
		)

		globalHost, globalDHT, errSetupLibP2P = ipfslite.SetupLibp2p(
			ctx,
			priv,
			nil,
			[]multiaddr.Multiaddr{sourceMultiAddr},
			nil,
			opts...,
		)
		if errSetupLibP2P != nil {
			return
		}

		globalBootstrapNodes, errSetupLibP2P = ethSmartContract.GetBoostrapNodes()
		if errSetupLibP2P != nil {
			return
		}
	})

	if errSetupLibP2P != nil {
		return nil, nil, errors.Wrap(errSetupLibP2P, "setup lib p2p")
	}

	return globalHost, globalDHT, nil
}
