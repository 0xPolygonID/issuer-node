package domain

import (
	"fmt"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/polygonid/sh-id-platform/internal/common"
	"strings"
	"time"
)

type CredentialAttributes struct {
	Name  string
	Value string
}

type LinkCoreDID core.DID

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

func (l *Link) IssuerCoreDID() *core.DID {
	return common.ToPointer(core.DID(l.IssuerDID))
}

func (l *Link) CredentialAttributesString() string {
	if len(l.CredentialAttributes) == 0 {
		return ""
	}

	credentialAttributesAsString := ""
	for i, v := range l.CredentialAttributes {
		credentialAttributesAsString += v.Name + ":" + v.Value
		if i != len(l.CredentialAttributes)-1 {
			credentialAttributesAsString += " "
		}
	}

	return credentialAttributesAsString
}

func (l *Link) StrToCredentialAttributes(schemaAttributess string) {
	l.CredentialAttributes = []CredentialAttributes{}
	attributes := strings.Split(schemaAttributess, " ")
	for _, attrElem := range attributes {
		attr := strings.Split(attrElem, ":")
		l.CredentialAttributes = append(l.CredentialAttributes, CredentialAttributes{
			Name:  attr[0],
			Value: attr[1],
		})
	}
}

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
