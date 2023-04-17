package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type logger func(ctx context.Context, msg string, args ...any)

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
	log  logger
}

// NewRedis returns a redis pubsub client
func NewRedis(rdb *redis.Client) *RedisClient {
	return &RedisClient{conn: rdb, log: func(ctx context.Context, msg string, args ...any) {}}
}

// WithLogger inject a function log that will be used from now on to log errors.
func (rdb *RedisClient) WithLogger(logFn logger) {
	rdb.log = logFn
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
					rdb.log(ctx, "redis pubsub: msg channel != topic")
					continue
				}
				if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
					rdb.log(ctx, "redis pubsub: unmarshalling payload event")
					continue
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							rdb.log(ctx, "panic in event handler", "r", r, "topic", topic)
						}
					}()
					if err := callback(ctx, payload.Msg); err != nil {
						rdb.log(ctx, "executing callback function", "err", err, "topic", topic)
					}
				}()

			case <-ctx.Done():
				return
			}
		}
	}()
}
