package pubsub

import (
	"context"
	"encoding/json"

	"github.com/mitchellh/mapstructure"
)

const (
	EventCreateCredential = "createCredential" // EventCreateCredential create credential event
)

// Event defines the payload
type Event interface{}

// CreateCredentialEvent defines the createCredential data
type CreateCredentialEvent struct {
	CredentialID string `json:"credentialID"`
	IssuerID     string `json:"issuerID"`
}

// MarshalBinary returns the bytes of an event
func (c CreateCredentialEvent) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

// UnmarshalEvent decodes the event into the to parameter
func UnmarshalEvent(from Event, to any) error {
	return mapstructure.Decode(from, to)
}

// Publisher sends topics to the pubsub
type Publisher interface {
	Publish(ctx context.Context, topic string, payload Event) error
}

// EventHandler is the type that functions that handle an event must comply.
type EventHandler func(context.Context, Event) error

// Subscriber subscribes to the pubsub topics
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, callback EventHandler)
}

// Client is formed by the publisher and subscriber
type Client interface {
	Publisher
	Subscriber
}
