package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// payload is the wrapper we will use to send pubsub messages to redis.
// ID is a unique identifier that could be used to ensure idempotency
// Time is the creation time of the event.
// Msg is the payload to send.
type payload struct {
	ID   uuid.UUID
	Time time.Time
	Msg  []byte
}

// MarshalBinary satisfies BinaryMarshaller interface and it is required by redis
func (p payload) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// RedisClient struct
type RedisClient struct {
	conn *redis.Client
}

// NewRedis returns a redis pubsub client
func NewRedis(rdb *redis.Client) Client {
	return &RedisClient{rdb}
}

// Publish publishes a new topic payload
func (rdb *RedisClient) Publish(ctx context.Context, topic string, event Event) error {
	msg, err := event.Marshal()
	if err != nil {
		return err
	}
	payload := payload{
		ID:   uuid.New(),
		Time: time.Now(),
		Msg:  []byte(msg),
	}
	return rdb.conn.Publish(ctx, topic, payload).Err()
}

// Subscribe adds a topic to the
func (rdb *RedisClient) Subscribe(ctx context.Context, topic string, callback EventHandler) {
	pubsub := rdb.conn.Subscribe(ctx, topic)
	go func() {
		var payload payload
		for {
			select {
			case event := <-pubsub.Channel():
				if event.Channel != topic {
					log.Error(ctx, "redis pubsub: msg channel != topic")
					continue
				}
				if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
					log.Error(ctx, "redis pubsub: unmarshalling payload event")
					continue
				}
				if err := callback(ctx, payload.Msg); err != nil {
					log.Error(ctx, "executing callback function", "err", err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
