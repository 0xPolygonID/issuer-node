package pubsub

import (
	"context"
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
}
