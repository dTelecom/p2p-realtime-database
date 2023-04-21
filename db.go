package p2p_database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-kad-dht/dual"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-datastore"
	ipfs_datastore "github.com/ipfs/go-datastore/sync"
	crdt "github.com/ipfs/go-ds-crdt"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultDatabaseEventsBufferSize = 128
	RebroadcastingInterval          = 1 * time.Second

	NetSubscriptionTopicPrefix  = "crdt_net_"
	NetSubscriptionPublishValue = "ping"
)

var (
	ErrEmptyKey = errors.New("empty key")
)

type PubSubHandler func(Event)

type TopicSubscription struct {
	subscription *pubsub.Subscription
	topic        *pubsub.Topic
	handler      PubSubHandler
}

type Event struct {
	Message string
}

type DB struct {
	Name   string
	selfID peer.ID
	host   host.Host
	crdt   *crdt.Datastore

	pubSub             *pubsub.PubSub
	topicSubscriptions map[string]*TopicSubscription
	handleGroup        *errgroup.Group
	lock               sync.RWMutex

	netTopic        *pubsub.Topic
	netSubscription *pubsub.Subscription
}

func Connect(
	ctx context.Context,
	h host.Host,
	dht *dual.DHT,
	bootstrapPeers []peer.AddrInfo,
	name string,
	opts ...dht.Option,
) (*DB, error) {
	crypto.MinRsaKeyBits = 1024

	grp := &errgroup.Group{}
	grp.SetLimit(DefaultDatabaseEventsBufferSize)

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, errors.Wrap(err, "create pubsub")
	}

	ds := ipfs_datastore.MutexWrap(datastore.NewMapDatastore())
	ipfs, err := ipfslite.New(ctx, ds, nil, h, dht, nil)
	if err != nil {
		return nil, errors.Wrap(err, "init ipfs")
	}
	ipfs.Bootstrap(bootstrapPeers)

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

	pubsubBC, err := crdt.NewPubSubBroadcaster(ctx, ps, "crdt_"+name)
	if err != nil {
		return nil, errors.Wrap(err, "init pub sub crdt broadcaster")
	}

	datastoreCrdt, err := crdt.New(ds, datastore.NewKey("crdt_"+name), ipfs, pubsubBC, crtdOpts)
	if err != nil {
		return nil, errors.Wrap(err, "init crdt")
	}

	err = datastoreCrdt.Sync(ctx, datastore.NewKey("/"))
	if err != nil {
		return nil, errors.Wrap(err, "crdt sync datastore")
	}

	netTopic, err := ps.Join(NetSubscriptionTopicPrefix + name)
	if err != nil {
		return nil, errors.Wrap(err, "create net topic")
	}

	netSubscription, err := netTopic.Subscribe()
	if err != nil {
		return nil, errors.Wrap(err, "subscribe to net topic")
	}

	db := &DB{
		Name:   name,
		host:   h,
		selfID: h.ID(),

		crdt: datastoreCrdt,

		pubSub:             ps,
		topicSubscriptions: map[string]*TopicSubscription{},
		handleGroup:        grp,
		lock:               sync.RWMutex{},

		netTopic:        netTopic,
		netSubscription: netSubscription,
	}

	db.refreshPeers(ctx)

	return db, nil
}

func (d *DB) List(ctx context.Context) ([]string, error) {
	r, err := d.crdt.Query(ctx, query.Query{KeysOnly: true})
	if err != nil {
		return nil, errors.Wrap(err, "crdt list query")
	}

	var keys []string
	for k := range r.Next() {
		keys = append(keys, k.Key)
	}

	return keys, nil
}

func (d *DB) Set(ctx context.Context, key, value string) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	err := d.crdt.Put(ctx, datastore.NewKey(key), []byte(value))
	if err != nil {
		return errors.Wrap(err, "crdt put")
	}

	return nil
}

func (d *DB) Get(ctx context.Context, key string) (string, error) {
	if len(key) == 0 {
		return "", ErrEmptyKey
	}

	val, err := d.crdt.Get(ctx, datastore.NewKey(key))
	if err != nil {
		return "", errors.Wrap(err, "crdt get")
	}

	return string(val), nil
}

func (d *DB) Remove(ctx context.Context, key string) error {
	err := d.crdt.Delete(ctx, datastore.NewKey(key))
	if err != nil {
		return errors.Wrap(err, "crdt delete key")
	}
	return nil
}

func (d *DB) Subscribe(ctx context.Context, topic string, handler PubSubHandler, opts ...pubsub.TopicOpt) error {
	t, err := d.joinTopic(topic, opts...)
	if err != nil {
		return err
	}

	s, err := t.Subscribe()
	if err != nil {
		return errors.Wrap(err, "pub sub subscribe topic")
	}

	d.lock.Lock()
	d.topicSubscriptions[topic] = &TopicSubscription{
		subscription: s,
		topic:        t,
	}
	d.lock.Unlock()

	d.handleGroup.Go(func() error {
		err = d.listenEvents(ctx, d.topicSubscriptions[topic])
		if err != nil {
			log.Log().Err(err).Str("topic", t.String()).Msg("pub sub listen events")
		}
		return err
	})

	return nil
}

func (d *DB) Publish(ctx context.Context, topic, value string, opts ...pubsub.PubOpt) error {
	t, err := d.joinTopic(topic)
	if err != nil {
		return err
	}

	err = t.Publish(ctx, []byte(value), opts...)
	if err != nil {
		return errors.Wrap(err, "pub sub publish message")
	}

	return nil
}

func (d *DB) Disconnect(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return d.crdt.Close()
	})

	g.Go(func() error {
		d.lock.RLock()
		for _, s := range d.topicSubscriptions {
			s.subscription.Cancel()
			err := s.topic.Close()
			if err != nil {
				log.Err(err).
					Str("current_peer_id", d.selfID.String()).
					Str("topic", s.topic.String()).
					Msg("try close db topic")
			}
		}
		d.lock.RUnlock()

		return nil
	})

	g.Go(func() error {
		return d.handleGroup.Wait()
	})

	return g.Wait()
}

func (d *DB) joinTopic(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	ts, ok := d.topicSubscriptions[topic]

	//already joined
	if ok {
		return ts.topic, nil
	}

	t, err := d.pubSub.Join(topic, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "pub sub join topic")
	}

	return t, nil
}

func (d *DB) listenEvents(ctx context.Context, topicSub *TopicSubscription) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := topicSub.subscription.Next(ctx)
			if err != nil {
				log.Err(err).
					Str("current_peer_id", d.selfID.String()).
					Msg("try get next pub sub message")

				continue
			}

			//skip self messages
			if msg.ReceivedFrom == d.selfID {
				continue
			}

			event := Event{}
			err = json.Unmarshal(msg.Data, &event)
			if err != nil {
				log.Err(err).
					Str("current_peer_id", d.selfID.String()).
					Str("from", msg.ReceivedFrom.String()).
					Str("message", string(msg.Data)).
					Msg("try unmarshal pub sub message")
			}

			topicSub.handler(event)
		}
	}
}

func (d *DB) refreshPeers(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := d.netSubscription.Next(ctx)
				if err != nil {
					log.Err(err).Msg("try net subscription read next message")
					continue
				}

				pubKey, err := msg.ReceivedFrom.ExtractPublicKey()
				if err != nil {
					log.Err(err).Str("peer_id", msg.ReceivedFrom.String()).Msg("cannot extract pub key from peer")
					continue
				}

				_, err = GetEthAddrFromPeer(pubKey)
				if err != nil {
					log.Err(err).Str("peer_id", msg.ReceivedFrom.String()).Msg("cannot extract eth addr from pub key")
					continue
				}

				//log.Info().Msg(ethAddr)

				if msg.ReceivedFrom == d.host.ID() {
					continue
				}

				d.host.ConnManager().TagPeer(msg.ReceivedFrom, "keep", 100)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := d.netTopic.Publish(ctx, []byte(NetSubscriptionPublishValue))
				if err != nil {
					log.Err(err).
						Str("current_peer_id", d.selfID.String()).
						Msg("try publish message to net ps topic")
				}
				time.Sleep(20 * time.Second)
			}
		}
	}()

}

func GetEthAddrFromPeer(pubkey crypto.PubKey) (string, error) {
	dbytes, _ := pubkey.Raw()
	k, err := secp256k1.ParsePubKey(dbytes)

	if err != nil {
		return "", err
	}

	return eth_crypto.PubkeyToAddress(*k.ToECDSA()).Hex(), nil
}
