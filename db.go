package p2p_database

import (
	"context"
	"encoding/json"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultDatabaseEventsBufferSize = 128
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
	kadDHT *dht.IpfsDHT

	pubSub             *pubsub.PubSub
	topicSubscriptions map[string]*TopicSubscription
	handleGroup        *errgroup.Group
	lock               sync.RWMutex
}

func Connect(
	ctx context.Context,
	h host.Host,
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
		Name:   name,
		selfID: h.ID(),

		kadDHT: kadDHT,

		pubSub:             ps,
		topicSubscriptions: map[string]*TopicSubscription{},
		handleGroup:        grp,
		lock:               sync.RWMutex{},
	}

	return db, nil
}

func (d *DB) Set(ctx context.Context, key, value string) error {
	err := d.kadDHT.PutValue(ctx, key, []byte(value))
	if err != nil {
		return errors.Wrap(err, "kad dht put value")
	}
	return nil
}

func (d *DB) Get(ctx context.Context, key string) (string, error) {
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
