package pubsub

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// RedisClient struct
type RedisClient struct {
	conn *redis.Client
}

// NewRedis returns a redis pubsub client
func NewRedis(rdb *redis.Client) Client {
	return &RedisClient{rdb}
}

// Publish publishes a new topic payload
func (rdb *RedisClient) Publish(ctx context.Context, topic string, payload Event) error {
	return rdb.conn.Publish(ctx, topic, payload).Err()
}

// Subscribe adds a topic to the
func (rdb *RedisClient) Subscribe(ctx context.Context, topic string, callback EventHandler) {
	pubsub := rdb.conn.Subscribe(ctx, topic)
	go func() {
		for {
			select {
			case event := <-pubsub.Channel():
				if event.Channel != topic {
					log.Error(ctx, "msg channel != topic")
					continue
				}

				var payload Event
				err := json.Unmarshal([]byte(event.Payload), &payload)
				if err != nil {
					log.Error(ctx, "unmarshal msg payload", "err", err)
					continue
				}

				err = callback(ctx, payload)
				if err != nil {
					log.Error(ctx, "executing callback function", "err", err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
