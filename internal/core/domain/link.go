package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/issuer-node/internal/common"
)

const (
	TypeString  = "string"  // TypeString is a string schema attribute type
	TypeInteger = "integer" // TypeInteger is an integer schema attribute type
	TypeBoolean = "boolean" // TypeBoolean is a boolean schema attribute type
)

// CredentialSubject holds a credential attribute item in string, string format as it is coming in the request.
type CredentialSubject map[string]interface{}

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
	Name     string
	Value    interface{}
	AttrType string
}

const (
	linkActive   = "active"   // LinkActive Link is active and can be used
	linkInactive = "inactive" // LinkInactive Link is inactive
	LinkExceeded = "exceeded" // LinkExceeded link usage exceeded.
)

// LinkCoreDID - represents a credential offer ID
type LinkCoreDID w3c.DID

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
	CredentialSubject        CredentialSubject
	Active                   bool
	Schema                   *Schema
	IssuedClaims             int // TODO: Give a value when link redemption is implemented
}

// NewLink - Constructor
func NewLink(
	issuerDID w3c.DID,
	maxIssuance *int,
	validUntil *time.Time,
	schemaID uuid.UUID,
	credentialExpiration *time.Time,
	credentialSignatureProof bool,
	credentialMTPProof bool,
	credentialSubject CredentialSubject,
) *Link {
	return &Link{
		ID:                       uuid.New(),
		IssuerDID:                LinkCoreDID(issuerDID),
		MaxIssuance:              maxIssuance,
		ValidUntil:               validUntil,
		SchemaID:                 schemaID,
		CredentialExpiration:     credentialExpiration,
		CredentialSignatureProof: credentialSignatureProof,
		CredentialMTPProof:       credentialMTPProof,
		CredentialSubject:        credentialSubject,
		Active:                   true,
		IssuedClaims:             0,
	}
}

// IssuerCoreDID - return the Core DID value
func (l *Link) IssuerCoreDID() *w3c.DID {
	return common.ToPointer(w3c.DID(l.IssuerDID))
}

// Scan - scan the value for LinkCoreDID
func (l *LinkCoreDID) Scan(value interface{}) error {
	didStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type, expected string")
	}
	did, err := w3c.ParseDID(didStr)
	if err != nil {
		return err
	}
	*l = LinkCoreDID(*did)
	return nil
}

// Status returns the status of the link based on the Active field, the number of issued claims or whether is expired or not
// If active is set to false, return "inactive"
// If maxIssuance is set and bypassed, returns "exceeded"
// If validUntil is set and bypassed, returns "exceeded"
// Otherwise return active.
func (l *Link) Status() string {
	if !l.Active {
		return linkInactive
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(time.Now()) {
		return LinkExceeded
	}
	if l.MaxIssuance != nil {
		if l.IssuedClaims >= *l.MaxIssuance {
			return LinkExceeded
		}
	}
	return linkActive
}
