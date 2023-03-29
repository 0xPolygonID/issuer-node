package link

import (
	"encoding/json"
	"fmt"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

const (
	// StatusPending - status pending
	StatusPending = "pending"
	// StatusError - status error
	StatusError = "error"
	// StatusDone - status done
	StatusDone = "done"
)

// State - Link state.
type State struct {
	Status  string                   `json:"status,omitempty"`
	Message string                   `json:"message,omitempty"`
	QRCode  *ports.LinkQRCodeMessage `json:"qrcode,omitempty"`
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
func NewStateDone(qrcode *ports.LinkQRCodeMessage) string {
	state := &State{
		Status: StatusDone,
		QRCode: qrcode,
	}
	return state.String()
}
