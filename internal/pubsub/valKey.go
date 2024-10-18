package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"

	"github.com/polygonid/sh-id-platform/internal/log"
)

type valkeyClient struct {
	client valkey.Client
}

// NewValKeyClient returns a new pubsub client based on Valkey
func NewValKeyClient(client valkey.Client) Client {
	return &valkeyClient{
		client: client,
	}
}

// Publish publishes a new topic payload
func (vk *valkeyClient) Publish(ctx context.Context, topic string, event Event) error {
	msg, err := event.Marshal()
	if err != nil {
		return err
	}
	payload := payload{
		ID:   uuid.New(),
		Time: time.Now(),
		Msg:  []byte(msg),
	}

	p, err := payload.MarshalBinary()
	if err != nil {
		log.Error(ctx, "error marshalling payload", "err:", err)
		return err
	}
	vk.client.Do(ctx, vk.client.B().Publish().Channel(topic).Message(string(p)).Build())
	return nil
}

// Subscribe adds a topic to the subscriber
func (vk *valkeyClient) Subscribe(ctx context.Context, topic string, callback EventHandler) {
	err := vk.client.Receive(ctx, vk.client.B().Subscribe().Channel(topic).Build(), func(msg valkey.PubSubMessage) {
		var payload payload
		err := json.Unmarshal([]byte(msg.Message), &payload)
		if err != nil {
			log.Error(ctx, "error unmarshalling payload", "err:", err)
			return
		}
		err = callback(ctx, payload.Msg)
		if err != nil {
			log.Error(ctx, "error processing message", "err:", err)
		}
	})
	if err != nil {
		log.Error(ctx, "error subscribing to topic", "err:", err)
	}
	<-ctx.Done()
}

// Close closes the pubsub client
func (vk *valkeyClient) Close() error {
	vk.client.Close()
	return nil
}
