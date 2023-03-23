package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// CredentialAttributes - credential's attributes
type CredentialAttributes struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LinkCoreDID - represents a credential offer ID
type LinkCoreDID core.DID

// Link - represents a credential offer
type Link struct {
	ID                       uuid.UUID
	IssuerDID                LinkCoreDID
	CreatedAt                time.Time
	MaxIssuance              *int32
	ValidUntil               *time.Time
	SchemaID                 uuid.UUID
	CredentialExpiration     *time.Time
	CredentialSignatureProof bool
	CredentialMTPProof       bool
	CredentialAttributes     []CredentialAttributes
	Active                   bool
}

// NewLink - Constructor
func NewLink(IssuerDID core.DID, maxIssuance *int32, validUntil *time.Time, schemaID uuid.UUID, CredentialSignatureProof bool, CredentialMTPProof bool, credentialAttributes []CredentialAttributes) *Link {
	return &Link{
		IssuerDID:                LinkCoreDID(IssuerDID),
		MaxIssuance:              maxIssuance,
		ValidUntil:               validUntil,
		SchemaID:                 schemaID,
		CredentialSignatureProof: CredentialSignatureProof,
		CredentialMTPProof:       CredentialMTPProof,
		CredentialAttributes:     credentialAttributes,
		Active:                   true,
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
