package pubsub

import (
	"context"

	"github.com/valkey-io/valkey-go"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/redis"
)

// Event defines the payload
type Event interface {
	Marshal() (msg Message, err error)
	Unmarshal(msg Message) error
}

// Message is the payload received in a pubsub subscriber. The input for callback functions
type Message []byte

// Publisher sends topics to the pubsub
type Publisher interface {
	Publish(ctx context.Context, topic string, payload Event) error
}

// EventHandler is the type that functions that handle an MyEvent must comply.
type EventHandler func(context.Context, Message) error

// Subscriber subscribes to the pubsub topics
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, callback EventHandler)
}

// Client is formed by the publisher and subscriber
type Client interface {
	Publisher
	Subscriber
	Close() error
}

// NewPubSub - creates a new pubsub client based on the configuration
func NewPubSub(ctx context.Context, cfg config.Configuration) (Client, error) {
	var ps Client
	if cfg.Cache.Provider == config.CacheProviderRedis {
		rdb, err := redis.Open(ctx, cfg.Cache.Url)
		if err != nil {
			log.Error(ctx, "cannot connect to redis", "err", err, "host", cfg.Cache.Url)
			return nil, err
		}
		ps = NewRedis(rdb)
	} else if cfg.Cache.Provider == config.CacheProviderValKey {
		client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{cfg.Cache.Url}})
		if err != nil {
			log.Error(ctx, "cannot connect to valkey", "err", err, "host", cfg.Cache.Url)
			return nil, err
		}
		ps = NewValKeyClient(client)
	}

	return ps, nil
}
