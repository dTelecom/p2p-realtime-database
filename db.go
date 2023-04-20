package p2p_database

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/go-cid"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultDatabaseEventsBufferSize = 128
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
	Name    string
	selfID  peer.ID
	host    host.Host
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
	kadDHT  *dht.IpfsDHT

	pubSub             *pubsub.PubSub
	topicSubscriptions map[string]*TopicSubscription
	handleGroup        *errgroup.Group
	lock               sync.RWMutex
}

func Connect(
	ctx context.Context,
	h host.Host,
	privKey crypto.PrivKey,
	pubKey crypto.PubKey,
	name string,
	ps *pubsub.PubSub,
	opts ...dht.Option,
) (*DB, error) {
	kadDHT, err := dht.New(ctx, h, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create dht")
	}

	err = kadDHT.Bootstrap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "bootstrap kad dht")
	}

	grp := &errgroup.Group{}
	grp.SetLimit(DefaultDatabaseEventsBufferSize)

	db := &DB{
		Name:    name,
		host:    h,
		privKey: privKey,
		pubKey:  pubKey,
		selfID:  h.ID(),

		kadDHT: kadDHT,

		pubSub:             ps,
		topicSubscriptions: map[string]*TopicSubscription{},
		handleGroup:        grp,
		lock:               sync.RWMutex{},
	}

	return db, nil
}

func (d *DB) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	pref := cid.Prefix{
		Version:  0,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	cast, err := pref.Sum([]byte(value))
	fmt.Println(cast)
	key = "/ipns/" + cast.String()
	fmt.Println(key)

	ipnsRecord, err := ipns.Create(d.privKey, []byte(value), 0, time.Now(), ttl)
	if err != nil {
		return errors.Wrap(err, "create IPNS record")
	}
	err = ipns.EmbedPublicKey(d.pubKey, ipnsRecord)
	if err != nil {
		return errors.Wrap(err, "embed pubkey to ipns record")
	}

	msg, err := proto.Marshal(ipnsRecord)
	if err != nil {
		return errors.Wrap(err, "marshal IPNS entry")
	}

	err = d.kadDHT.PutValue(ctx, key, msg)
	if err != nil {
		return errors.Wrap(err, "kad dht put value")
	}
	return nil
}

func (d *DB) Get(ctx context.Context, key string) (string, error) {
	if len(key) == 0 {
		return "", ErrEmptyKey
	}

	if key[0] != '/' {
		key = "/" + d.Name + "/" + key
	}

	b, err := d.kadDHT.GetValue(ctx, key)
	if err != nil {
		return "", errors.Wrap(err, "kad dht get value")
	}
	return string(b), nil
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

		return d.kadDHT.Close()
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
