package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
)

const (
	TypeString  = "string"  // TypeString is a string schema attribute type
	TypeInteger = "integer" // TypeInteger is an integer schema attribute type
	TypeBoolean = "boolean" // TypeBoolean is a boolean schema attribute type
)

// CredentialAttrsRequest holds a credential attribute item in string, string format as it is coming in the request.
type CredentialAttrsRequest struct {
	Name  string
	Value string
}

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

// NewLink - Constructor
func NewLink(
	issuerDID core.DID,
	maxIssuance *int,
	validUntil *time.Time,
	schemaID uuid.UUID,
	credentialExpiration *time.Time,
	CredentialSignatureProof bool,
	CredentialMTPProof bool,
) *Link {
	return &Link{
		ID:                       uuid.New(),
		IssuerDID:                LinkCoreDID(issuerDID),
		MaxIssuance:              maxIssuance,
		ValidUntil:               validUntil,
		SchemaID:                 schemaID,
		CredentialExpiration:     credentialExpiration,
		CredentialSignatureProof: CredentialSignatureProof,
		CredentialMTPProof:       CredentialMTPProof,
		CredentialAttributes:     make([]CredentialAttributes, 0),
		Active:                   true,
		IssuedClaims:             0,
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

// LoadAttributeTypes uses the remote schema to add attribute types to the list of CredentialAttributes.
// This is needed to load data from the DB because we are not storing the attribute types on it.
func (l *Link) LoadAttributeTypes(ctx context.Context, ld loader.Loader) error {
	schema, err := jsonschema.Load(ctx, ld)
	if err != nil {
		return err
	}
	for i, credentialAttr := range l.CredentialAttributes {
		attrData, err := schema.AttributeByID(credentialAttr.Name)
		if err != nil {
			return err
		}
		l.CredentialAttributes[i].AttrType = attrData.Type
	}
	return nil
}

// ProcessAttributes - validates an array of attributes against the schema and populates field CredentialAttributes
func (l *Link) ProcessAttributes(ctx context.Context, ld loader.Loader, attrs []CredentialAttrsRequest) error {
	schema, err := jsonschema.Load(ctx, ld)
	if err != nil {
		return err
	}
	schemaAttributes, err := schema.Attributes()
	if err != nil {
		return fmt.Errorf("processing schema: %w", err)
	}
	if len(schemaAttributes) != (len(attrs) + 1) { // +1 because of extra attr @context in json credential files
		return fmt.Errorf("the number of attributes is not valid")
	}
	for _, attr := range attrs {
		attrData, err := schema.AttributeByID(attr.Name)
		if err != nil {
			return err
		}
		value, aType, err := validateCredentialLinkAttribute(*attrData, attr.Name, attr.Value)
		if err != nil {
			return err
		}
		l.CredentialAttributes = append(l.CredentialAttributes, CredentialAttributes{
			Name:     attr.Name,
			Value:    value,
			AttrType: aType,
		})
	}
	return nil
}

func validateCredentialLinkAttribute(attr jsonschema.Attribute, name string, val string) (interface{}, string, error) {
	switch attr.Type {
	case TypeString:
		return val, TypeString, nil
	case TypeInteger:
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, TypeInteger, fmt.Errorf("converting attribute <%s> :%w", name, err)
		}
		return i, TypeInteger, nil
	case TypeBoolean:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, TypeBoolean, fmt.Errorf("converting attribute <%s> :%w", name, err)
		}
		return b, TypeBoolean, nil
	}
	return nil, attr.Type, fmt.Errorf("converting attribute <%s>", name)
}
