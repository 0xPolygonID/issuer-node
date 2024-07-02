package event

import (
	"encoding/json"

	"github.com/polygonid/sh-id-platform/pkg/pubsub"
)

const (
	CreateCredentialEvent = "createCredentialEvent" // CreateCredentialEvent create credential event
	CreateConnectionEvent = "createConnectionEvent" // CreateConnectionEvent create connection MyEvent
	CreateStateEvent      = "createStateEvent"      // CreateStateEvent create state event
)

// CreateState defines the createState data
type CreateState struct {
	State string `json:"state"`
}

// Marshal marshals the event into a pubsub.Message
func (ev *CreateState) Marshal() (msg pubsub.Message, err error) {
	return json.Marshal(ev)
}

// Unmarshal creates an event from that message
func (ev *CreateState) Unmarshal(msg pubsub.Message) error {
	return json.Unmarshal(msg, &ev)
}

// CreateCredential defines the createCredential data
type CreateCredential struct {
	CredentialIDs []string `json:"credentialsID"`
	IssuerID      string   `json:"issuerID"`
}

// Marshal marshals the event into a pubsub.Message
func (ev *CreateCredential) Marshal() (msg pubsub.Message, err error) {
	return json.Marshal(ev)
}

// Unmarshal creates an event from that message
func (ev *CreateCredential) Unmarshal(msg pubsub.Message) error {
	return json.Unmarshal(msg, &ev)
}

// CreateConnection defines the createCredential data
type CreateConnection struct {
	ConnectionID string `json:"connectionID"`
	IssuerID     string `json:"issuerID"`
}

// Marshal marshals the event into a pubsub.Message
func (ev *CreateConnection) Marshal() (msg pubsub.Message, err error) {
	return json.Marshal(ev)
}

// Unmarshal creates an event from that message
func (ev *CreateConnection) Unmarshal(msg pubsub.Message) error {
	return json.Unmarshal(msg, &ev)
}
