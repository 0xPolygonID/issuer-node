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
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

const (
	linkActive   = "active"   // LinkActive Link is active and can be used
	linkInactive = "inactive" // LinkInactive Link is inactive
	LinkExceed   = "exceed"   // LinkExceed link usage exceeded.
)

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
	Schema                   *Schema
	IssuedClaims             int // TODO: Give a value when link redemption is implemented
}

// Status returns the status of the link based on the Active field, the number of issued claims or whether is expired or not
// If active is set to false, return "inactive"
// If maxIssuance is set and bypassed, returns "exceed"
// If validUntil is set and bypassed, returns ""exceed"
// Otherwise return active.
func (l *Link) Status() string {
	if !l.Active {
		return linkInactive
	}
	if l.ValidUntil != nil && l.ValidUntil.Before(time.Now()) {
		return LinkExceed
	}
	if l.MaxIssuance != nil {
		if l.IssuedClaims == *l.MaxIssuance {
			return linkInactive
		}
		if l.IssuedClaims > *l.MaxIssuance {
			return LinkExceed
		}
	}
	return linkActive
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
