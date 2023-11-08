package link

import (
	"encoding/json"
	"fmt"
)

const (
	// StatusPending - status pending
	StatusPending = "pending"
	// StatusPendingPublish - status StatusPendingPublish
	StatusPendingPublish = "pendingPublish"
	// StatusError - status error
	StatusError = "error"
	// StatusDone - status done
	StatusDone = "done"
)

// CredentialOfferMessageType - TODO
const CredentialOfferMessageType string = "https://iden3-communication.io/credentials/1.0/offer"

// CredentialLink is structure to fetch credential
type CredentialLink struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// CredentialsLinkMessageBody is struct the represents offer message
type CredentialsLinkMessageBody struct {
	URL         string           `json:"url"`
	Credentials []CredentialLink `json:"credentials"`
}

// QRCodeMessage - TODO
type QRCodeMessage struct {
	ID       string                     `json:"id"`
	Typ      string                     `json:"typ,omitempty"`
	Type     string                     `json:"type"`
	ThreadID string                     `json:"thid,omitempty"`
	Body     CredentialsLinkMessageBody `json:"body,omitempty"`
	From     string                     `json:"from,omitempty"`
	To       string                     `json:"to,omitempty"`
}

// State - Link state.
type State struct {
	Status  string  `json:"status,omitempty"`
	Message string  `json:"message,omitempty"`
	QRCode  *string `json:"qrcode,omitempty"`
}

// NewStatePending - TODO
func NewStatePending() *State {
	return &State{Status: StatusPending}
}

func (ls *State) String() string {
	s, _ := json.Marshal(ls)
	return string(s)
}

// CredentialStateCacheKey - TODO
func CredentialStateCacheKey(linkID, sessionID string) string {
	return fmt.Sprintf("credential_link_%s_%s", linkID, sessionID)
}

// NewStateError - NewStateError
func NewStateError(err error) *State {
	return &State{
		Status:  StatusError,
		Message: err.Error(),
	}
}

// NewStateDone - TODO
func NewStateDone(qrCodeLink string) *State {
	state := &State{
		Status: StatusDone,
		QRCode: &qrCodeLink,
	}
	return state
}

// NewStatePendingPublish - TODO
func NewStatePendingPublish() *State {
	state := &State{
		Status: StatusPendingPublish,
	}
	return state
}
