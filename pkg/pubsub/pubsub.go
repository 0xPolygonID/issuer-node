package pubsub

import (
	"context"
	"encoding/json"
)

const (
	EventCreateCredential = "createCredential" // EventCreateCredential create credential event
	EventCreateConnection = "createConnection" // EventCreateConnection create connection MyEvent
)

// Event defines the payload
type Event interface {
	Marshal() (msg Message, err error)
	Unmarshal(msg Message) error
}

// Message is the payload received in a pubsub subscriber. The input for callback functions
type Message []byte

// TODO: Move events to another internal package. They are not related to the library

// CreateCredentialEvent defines the createCredential data
type CreateCredentialEvent struct {
	CredentialID string `json:"credentialID"`
	IssuerID     string `json:"issuerID"`
}

// Marshal marshals the event into a pubsub.Message
func (ev *CreateCredentialEvent) Marshal() (msg Message, err error) {
	return json.Marshal(ev)
}

// Unmarshal creates an event from that message
func (ev *CreateCredentialEvent) Unmarshal(msg Message) error {
	return json.Unmarshal(msg, &ev)
}

// CreateConnectionEvent defines the createCredential data
type CreateConnectionEvent struct {
	ConnectionID string `json:"connectionID"`
	IssuerID     string `json:"issuerID"`
}

// Marshal marshals the event into a pubsub.Message
func (ev *CreateConnectionEvent) Marshal() (msg Message, err error) {
	return json.Marshal(ev)
}

// Unmarshal creates an event from that message
func (ev *CreateConnectionEvent) Unmarshal(msg Message) error {
	return json.Unmarshal(msg, &ev)
}

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
