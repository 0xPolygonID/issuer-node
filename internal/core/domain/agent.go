package domain

import "github.com/iden3/iden3comm"

// Agent struct
type Agent struct {
	ID       string                    `json:"id"`
	Typ      iden3comm.MediaType       `json:"typ,omitempty"`
	Type     iden3comm.ProtocolMessage `json:"type"`
	ThreadID string                    `json:"thid,omitempty"`
	Body     interface{}               `json:"body,omitempty"`
	From     string                    `json:"from,omitempty"`
	To       string                    `json:"to,omitempty"`
}
