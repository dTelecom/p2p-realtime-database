package p2p_database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	zaploki "github.com/paul-milne/zap-loki"
	"go.uber.org/zap"

	"github.com/google/uuid"
	ipfs_datastore "github.com/ipfs/go-datastore/sync"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-kad-dht/dual"

	ipfslite "github.com/hsanjuan/ipfs-lite"
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
	RebroadcastingInterval          = 30 * time.Second

	cleanupExpiredRecordsInterval = 1 * time.Second

	NetSubscriptionTopicPrefix  = "crdt_net_"
	NetSubscriptionPublishValue = "ping"

	TTLSubscribeTopicPrefix = "ttl_"
)

const DiscoveryTag = "p2p-database-discovery"

var (
	ErrEmptyKey                     = errors.New("empty key")
	ErrKeyNotFound                  = errors.New("key not found")
	ErrIncorrectSubscriptionHandler = errors.New("incorrect subscription handler")
)

var (
	alreadyConnectedLock = sync.RWMutex{}
	alreadyConnected     = make(map[peer.ID]multiaddr.Multiaddr)

	onceInitHostP2P = sync.Once{}
	lock            = sync.RWMutex{}

	globalPubSubCrdtBroadcasters = map[string]*crdt.PubSubBroadcaster{}
	globalHost                   host.Host
	globalDHT                    *dual.DHT
	globalBootstrapNodes         []peer.AddrInfo
	globalGossipSub              *pubsub.PubSub

	globalConnectionManager *ConnectionManager

	globalLockIPFS  = sync.RWMutex{}
	onceInitIPFS    = sync.Once{}
	globalReadyIPFS bool
	globalIPFs      *ipfslite.Peer

	globalDataStorePerDb          = map[string]datastore.Batching{}
	globalJoinedTopicsPerDb       = map[string]map[string]*pubsub.Topic{}
	globalTopicSubscriptionsPerDb = map[string]map[string]*TopicSubscription{}
)

type TTLMessage struct {
	Key         string    `json:"key"`
	RemoveAfter time.Time `json:"remove_after"`
}

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
	Name   string
	selfID peer.ID
	host   host.Host
	crdt   *crdt.Datastore

	ds          datastore.Batching
	pubSub      *pubsub.PubSub
	handleGroup *errgroup.Group
	lock        sync.RWMutex

	cancel context.CancelFunc

	ttlLock     sync.RWMutex
	ttlMessages map[string]TTLMessage

	ready             bool
	readyDatabaseLock sync.Mutex
	disconnectOnce    sync.Once

	logger       *logging.ZapEventLogger
	remoteLogger *zap.Logger
}

func Connect(
	ctx context.Context,
	config Config,
	logger *logging.ZapEventLogger,
	opts ...dht.Option,
) (*DB, error) {
	ctx, cancel := context.WithCancel(ctx)

	crypto.MinRsaKeyBits = 1024

	grp := &errgroup.Group{}
	grp.SetLimit(DefaultDatabaseEventsBufferSize)

	port := config.PeerListenPort
	if port == 0 {
		port = DefaultPort
	}
	h, _, err := makeHost(ctx, config, port)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "make lib p2p host")
	}

	lock.RLock()
	pubsubTopic := "crdt_" + config.DatabaseName
	pubsubBC, exists := globalPubSubCrdtBroadcasters[pubsubTopic]
	lock.RUnlock()

	if !exists {
		lock.Lock()
		pubsubBC, err = crdt.NewPubSubBroadcaster(context.Background(), globalGossipSub, pubsubTopic)
		if err != nil && !strings.Contains(err.Error(), "topic already exists") {
			cancel()
			return nil, errors.Wrap(err, "init pub sub crdt broadcaster")
		}
		globalPubSubCrdtBroadcasters[pubsubTopic] = pubsubBC
		lock.Unlock()
	}

	lock.RLock()
	ds, exists := globalDataStorePerDb[config.DatabaseName]
	lock.RUnlock()

	if !exists {
		lock.Lock()
		ds = ipfs_datastore.MutexWrap(datastore.NewMapDatastore())
		globalDataStorePerDb[config.DatabaseName] = ds
		lock.Unlock()
	}

	for i, bootstrapNode := range globalBootstrapNodes {
		logger.Infof("Bootstrap node %d - %s - [%s]", i, bootstrapNode.String(), bootstrapNode.Addrs[0].String())
		h.ConnManager().TagPeer(bootstrapNode.ID, "keep", 100)
	}

	crtdOpts := crdt.DefaultOptions()
	crtdOpts.Logger = logging.Logger("p2p_database_" + config.DatabaseName)
	crtdOpts.RebroadcastInterval = RebroadcastingInterval
	crtdOpts.PutHook = func(k datastore.Key, v []byte) {
		fmt.Printf("[%s] Added: [%s] -> %s\n", time.Now().Format(time.RFC3339), k, string(v))
		if config.NewKeyCallback != nil {
			config.NewKeyCallback(k.String())
		}
	}
	crtdOpts.DeleteHook = func(k datastore.Key) {
		fmt.Printf("[%s] Removed: [%s]\n", time.Now().Format(time.RFC3339), k)
		if config.RemoveKeyCallback != nil {
			config.RemoveKeyCallback(k.String())
		}
	}
	crtdOpts.RebroadcastInterval = time.Second

	doneBootstrappingIPFS, err := makeIPFS(ctx, ds, h)
	if err != nil {
		cancel()
		return nil, err
	}

	datastoreCrdt, err := crdt.New(ds, datastore.NewKey("crdt_"+config.DatabaseName), globalIPFs, pubsubBC, crtdOpts)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "init crdt")
	}

	err = datastoreCrdt.Sync(ctx, datastore.NewKey("/"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "crdt sync datastore")
	}

	lock.Lock()
	_, ok := globalJoinedTopicsPerDb[config.DatabaseName]
	if !ok {
		globalJoinedTopicsPerDb[config.DatabaseName] = map[string]*pubsub.Topic{}
	}
	_, ok = globalTopicSubscriptionsPerDb[config.DatabaseName]
	if !ok {
		globalTopicSubscriptionsPerDb[config.DatabaseName] = map[string]*TopicSubscription{}
	}
	lock.Unlock()

	var remoteLogger *zap.Logger
	if config.LokiLoggingHost != "" {
		zapConfig := zap.NewProductionConfig()
		zapConfig.OutputPaths = []string{"/dev/null"}

		loki := zaploki.New(context.Background(), zaploki.Config{
			Url:          config.LokiLoggingHost,
			BatchMaxSize: 1,
			BatchMaxWait: 10 * time.Second,
			Labels:       map[string]string{"database_name": config.DatabaseName, "peer_id": h.ID().String()},
		})
		remoteLogger, err = loki.WithCreateLogger(zapConfig)
		go func() {
			for {
				remoteLogger.Sync()
				time.Sleep(5 * time.Second)
			}
		}()
		if err != nil {
			logger.Errorf("try init remote logger with loki host %s: %s", config.LokiLoggingHost, err.Error())
		}
	}

	db := &DB{
		Name:   config.DatabaseName,
		host:   h,
		selfID: h.ID(),

		logger:       logger,
		remoteLogger: remoteLogger,

		crdt: datastoreCrdt,

		cancel:      cancel,
		ttlLock:     sync.RWMutex{},
		ttlMessages: map[string]TTLMessage{},

		pubSub:      globalGossipSub,
		handleGroup: grp,
		lock:        sync.RWMutex{},

		readyDatabaseLock: sync.Mutex{},
		disconnectOnce:    sync.Once{},
	}

	globalLockIPFS.RLock()
	if !globalReadyIPFS {
		//Unlock db after successfully bootstrapping IPFS
		db.readyDatabaseLock.Lock()
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-doneBootstrappingIPFS:
				db.readyDatabaseLock.Unlock()
				return
			}
		}()
	}
	globalLockIPFS.RUnlock()

	db.Subscribe(ctx, NetSubscriptionTopicPrefix+config.DatabaseName, func(event Event) {
		peerId, err := peer.Decode(event.FromPeerId)
		if err != nil {
			db.logger.Errorf("net topic parse peer %s error: %s", event.FromPeerId, err)
			if db.remoteLogger != nil {
				db.remoteLogger.Error(fmt.Sprintf("net topic parse peer %s error: %s", event.FromPeerId, err))
			}
			return
		}

		if event.FromPeerId == db.host.ID().String() {
			return
		}

		db.host.ConnManager().TagPeer(peerId, "keep", 100)
	})

	db.Subscribe(ctx, TTLSubscribeTopicPrefix+config.DatabaseName, func(event Event) {
		body := event.Message.(string)
		ttl := TTLMessage{}

		err = json.Unmarshal([]byte(body), &ttl)
		if err != nil {
			logger.Errorw("ttl topic handler: unmarshal body to TTLMessage "+body, err)
			return
		}

		db.ttlLock.Lock()
		db.ttlMessages[ttl.Key] = ttl
		db.ttlLock.Unlock()
	})

	db.netPingPeers(ctx, NetSubscriptionTopicPrefix+config.DatabaseName)
	db.startDiscovery(ctx)
	db.startCleanupExpiredTTLKeysProcess(ctx)

	go func() {
		if remoteLogger == nil {
			return
		}

		for {
			peers := db.ConnectedPeers()
			for _, p := range peers {
				remoteLogger.Info("connected peer", zap.String("peer", p.String()))
			}
			time.Sleep(15 * time.Second)
		}
	}()

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
	db.WaitReady(ctx)

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
	db.WaitReady(ctx)

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
	db.WaitReady(ctx)

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
	db.WaitReady(ctx)

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
	db.WaitReady(ctx)

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

	lock.Lock()
	globalTopicSubscriptionsPerDb[db.Name][topic] = &TopicSubscription{
		subscription: s,
		topic:        t,
		handler:      handler,
	}
	lock.Unlock()

	db.handleGroup.Go(func() error {
		err = db.listenEvents(ctx, globalTopicSubscriptionsPerDb[db.Name][topic])
		if err != nil {
			db.logger.Errorf("pub sub listen events topic %s err %s", topic, err)
		}
		return err
	})

	return nil
}

func (db *DB) Publish(ctx context.Context, topic string, value interface{}, opts ...pubsub.PubOpt) (Event, error) {
	db.WaitReady(ctx)

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

	if db.remoteLogger != nil {
		db.remoteLogger.Info("Send message to topic", zap.ByteString("message", marshaled), zap.String("topic", topic), zap.String("id", event.ID))
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
		db.cancel()
		err = db.disconnect(ctx)
	})
	return err
}

func (db *DB) disconnect(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		lock.Lock()
		defer lock.Unlock()

		topics := globalTopicSubscriptionsPerDb[db.Name]
		delete(globalTopicSubscriptionsPerDb, db.Name)
		delete(globalJoinedTopicsPerDb, db.Name)

		for _, s := range topics {
			s.subscription.Cancel()
			err := s.topic.Close()
			if err != nil {
				db.logger.Errorf("try close db topic %s current peer id %s: %s", s.topic, db.host.ID(), err)
			}
		}
		return nil
	})
	g.Go(func() error {
		err := db.handleGroup.Wait()
		return err
	})
	//g.Go(func() error {
	//	err := db.crdt.Close()
	//	if err != nil {
	//		return errors.Wrap(err, "crdt close")
	//	}
	//	return nil
	//})

	ch := make(chan error, 1)
	go func() {
		ch <- g.Wait()
	}()

	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "try close")
	case err := <-ch:
		return err
	}
}

func (db *DB) GetHost() host.Host {
	return db.host
}

func (db *DB) ConnectedPeers() []*peer.AddrInfo {
	var pinfos []*peer.AddrInfo
	for _, c := range db.host.Network().Conns() {
		pinfos = append(pinfos, &peer.AddrInfo{
			ID:    c.RemotePeer(),
			Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
		})
	}
	return pinfos
}

func (db *DB) joinTopic(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	lock.Lock()
	defer lock.Unlock()

	ts, ok := globalTopicSubscriptionsPerDb[db.Name][topic]
	//already joined
	if ok {
		return ts.topic, nil
	}

	if t, ok := globalJoinedTopicsPerDb[db.Name][topic]; ok {
		return t, nil
	}

	t, err := db.pubSub.Join(db.Name+"_"+topic, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "pub sub join topic")
	}
	globalJoinedTopicsPerDb[db.Name][topic] = t

	if db.remoteLogger != nil {
		db.remoteLogger.Info("subscribe to topic", zap.String("topic", topic))
	}

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
				if errors.Is(err, pubsub.ErrSubscriptionCancelled) || errors.Is(err, context.Canceled) {
					return nil
				}
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

			if db.remoteLogger != nil {
				db.remoteLogger.Info(
					"got message from topic",
					zap.String("topic", *msg.Topic),
					zap.String("id", event.ID),
					zap.String("from_id", event.FromPeerId),
					zap.ByteString("data", msg.Data),
				)
			}

			topicSub.handler(event)
		}
	}
}

func (db *DB) netPingPeers(ctx context.Context, netTopic string) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := db.Publish(ctx, netTopic, []byte(NetSubscriptionPublishValue))
				if err != nil {
					db.logger.Errorf("try publish message to net ps topic: %s", err)
					if errors.Is(err, pubsub.ErrTopicClosed) {
						return
					}
				}
				time.Sleep(20 * time.Second)
			}
		}
	}()
}

func (db *DB) TTL(ctx context.Context, key string, ttl time.Duration) error {
	ttlMessage := TTLMessage{
		Key:         key,
		RemoveAfter: time.Now().UTC().Add(ttl),
	}

	body, err := json.Marshal(ttlMessage)
	if err != nil {
		return errors.Wrap(err, "marshal ttl message")
	}

	db.ttlLock.Lock()
	db.ttlMessages[key] = ttlMessage
	db.ttlLock.Unlock()

	//notify another nodes about new ttl for key
	_, err = db.Publish(ctx, TTLSubscribeTopicPrefix+db.Name, string(body))
	if err != nil {
		return errors.Wrap(err, "publish message")
	}

	return nil
}

func (db *DB) startCleanupExpiredTTLKeysProcess(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(cleanupExpiredRecordsInterval)
		for {
			select {
			case <-ticker.C:
				var delayedKeysForDeletion []string

				db.ttlLock.RLock()
				for key, ttl := range db.ttlMessages {
					now := time.Now().UTC()
					removeAfter := ttl.RemoveAfter.UTC()

					if now.Before(removeAfter) {
						continue
					}

					lag := now.UnixMilli() - removeAfter.UnixMilli()
					if lag > int64(float64(cleanupExpiredRecordsInterval.Milliseconds())*1.5) {
						db.logger.Warnf("node lag %dms for remove expired record. forgot ttl for key %s", lag, key)

						db.ttlLock.RUnlock()
						db.ttlLock.Lock()
						delete(db.ttlMessages, key)
						db.ttlLock.Unlock()

						continue
					}

					delayedKeysForDeletion = append(delayedKeysForDeletion, ttl.Key)
				}
				db.ttlLock.RUnlock()

				for _, delayedKeyForDeletion := range delayedKeysForDeletion {
					err := db.Remove(ctx, delayedKeyForDeletion)
					if err != nil && !errors.Is(err, ErrKeyNotFound) {
						db.logger.Errorw("remove expired key", err)
					}

					db.ttlLock.Lock()
					delete(db.ttlMessages, delayedKeyForDeletion)
					db.ttlLock.Unlock()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (db *DB) WaitReady(ctx context.Context) {
	if db.ready || globalReadyIPFS {
		return
	}

	db.readyDatabaseLock.Lock()
	db.ready = true
	db.readyDatabaseLock.Unlock()
}

func (db *DB) startDiscovery(ctx context.Context) {
	db.WaitReady(ctx)

	rendezvous := DiscoveryTag + "_" + db.Name
	routingDiscovery := routing.NewRoutingDiscovery(globalDHT)
	util.Advertise(ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, err := util.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				db.logger.Errorf("discrovery find peers error %s", err)
				return
			}
			for _, p := range peers {
				db.logger.Errorf("found peer %s", p.String())

				if p.ID == db.host.ID() {
					continue
				}

				if db.host.Network().Connectedness(p.ID) != network.Connected {
					_, err = db.host.Network().DialPeer(ctx, p.ID)
					if err != nil {
						db.logger.Errorf("discrovery connected to peer error %s: %s", p.ID.String(), err)
						continue
					}
					db.logger.Infof("discrovery connected to peer %s\n", p.ID.String())
				}
			}
		}
	}()
}

func makeHost(ctx context.Context, config Config, port int) (host.Host, *dual.DHT, error) {
	ethSmartContract, err := NewEthSmartContract(config, logging.Logger("eth-smart-contract"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "create ethereum smart contract")
	}

	prvKey, err := eth_crypto.HexToECDSA(config.WalletPrivateKey)
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
		if !config.DisableGater {
			opts = append(opts, libp2p.ConnectionGater(
				NewEthConnectionGater(ethSmartContract, globalConnectionManager, *logging.Logger("eth-connection-gater")),
			))
		}

		globalHost, globalDHT, errSetupLibP2P = ipfslite.SetupLibp2p(
			context.Background(),
			priv,
			nil,
			[]multiaddr.Multiaddr{sourceMultiAddr},
			nil,
			opts...,
		)
		if errSetupLibP2P != nil {
			return
		}

		globalConnectionManager = NewConnectionManager(globalHost, logging.Logger("connection-manager"))

		globalGossipSub, errSetupLibP2P = pubsub.NewGossipSub(context.Background(), globalHost)
		if err != nil {
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

func makeIPFS(ctx context.Context, ds datastore.Batching, h host.Host) (chan struct{}, error) {
	var (
		err               error
		doneBootstrapping = make(chan struct{}, 1)
	)

	onceInitIPFS.Do(func() {
		globalIPFs, err = ipfslite.New(context.Background(), ds, nil, h, globalDHT, &ipfslite.Config{})
		go func() {
			globalIPFs.Bootstrap(globalBootstrapNodes)

			doneBootstrapping <- struct{}{}
			globalLockIPFS.Lock()
			globalReadyIPFS = true
			globalLockIPFS.Unlock()
		}()
	})

	return doneBootstrapping, errors.Wrap(err, "init ipfs")
}
