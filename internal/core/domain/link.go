package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// LinkRequestMessageMessageBody - TODO
type LinkRequestMessageMessageBody struct {
	CallbackURL string                               `json:"callbackUrl"`
	Reason      string                               `json:"reason,omitempty"`
	Message     string                               `json:"message,omitempty"`
	DIDDoc      json.RawMessage                      `json:"did_doc,omitempty"`
	Scope       []protocol.ZeroKnowledgeProofRequest `json:"scope"`
}

// LinkRequestMessage - TODO
type LinkRequestMessage struct {
	ID       string                        `json:"id"`
	Typ      iden3comm.MediaType           `json:"typ,omitempty"`
	Type     iden3comm.ProtocolMessage     `json:"type"`
	ThreadID string                        `json:"thid,omitempty"`
	Body     LinkRequestMessageMessageBody `json:"body,omitempty"`

	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// CredentialAttributes - credential's attributes
type CredentialAttributes struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// LinkCoreDID - represents a credential offer ID
type LinkCoreDID core.DID

// Link - represents a credential offer
type Link struct {
	ID                       uuid.UUID
	IssuerDID                LinkCoreDID
	CreatedAt                time.Time
	MaxIssuance              *int
	ValidUntil               *time.Time
	SchemaID                 uuid.UUID
	CredentialExpiration     *time.Time
	CredentialSignatureProof bool
	CredentialMTPProof       bool
	CredentialAttributes     []CredentialAttributes
	Active                   bool
	Issued                   int
}

// NewLink - Constructor
func NewLink(issuerDID core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID, credentialExpiration *time.Time, CredentialSignatureProof bool, CredentialMTPProof bool, credentialAttributes []CredentialAttributes) *Link {
	return &Link{
		ID:                       uuid.New(),
		IssuerDID:                LinkCoreDID(issuerDID),
		MaxIssuance:              maxIssuance,
		ValidUntil:               validUntil,
		SchemaID:                 schemaID,
		CredentialExpiration:     credentialExpiration,
		CredentialSignatureProof: CredentialSignatureProof,
		CredentialMTPProof:       CredentialMTPProof,
		CredentialAttributes:     credentialAttributes,
		Active:                   true,
		Issued:                   0,
	}
}

// IssuerCoreDID - return the Core DID value
func (l *Link) IssuerCoreDID() *core.DID {
	return common.ToPointer(core.DID(l.IssuerDID))
}

// Scan - scan the value for LinkCoreDID
func (l *LinkCoreDID) Scan(value interface{}) error {
	didStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type, expected string")
	}
	did, err := core.ParseDID(didStr)
	if err != nil {
		return err
	}
	*l = LinkCoreDID(*did)
	return nil
}
