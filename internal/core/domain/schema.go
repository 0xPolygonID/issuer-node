package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
)

//nolint:gosec //reason: constant
const (
	AuthBJJCredential              = "AuthBJJCredential"
	AuthBJJCredentialJSONSchemaURL = "https://schema.iden3.io/core/json/auth.json"
	AuthBJJCredentialSchemaJSON    = `{"$schema":"http://json-schema.org/draft-07/schema#","$metadata":{"uris":{"jsonLdContext":"https://schema.iden3.io/core/jsonld/auth.jsonld","jsonSchema":"https://schema.iden3.io/core/json/auth.json"},"serialization":{"indexDataSlotA":"x","indexDataSlotB":"y"}},"type":"object","required":["@context","id","type","issuanceDate","credentialSubject","credentialSchema","credentialStatus","issuer"],"properties":{"@context":{"type":["string","array","object"]},"id":{"type":"string"},"type":{"type":["string","array"],"items":{"type":"string"}},"issuer":{"type":["string","object"],"format":"uri","required":["id"],"properties":{"id":{"type":"string","format":"uri"}}},"issuanceDate":{"type":"string","format":"date-time"},"expirationDate":{"type":"string","format":"date-time"},"credentialSchema":{"type":"object","required":["id","type"],"properties":{"id":{"type":"string","format":"uri"},"type":{"type":"string"}}},"credentialSubject":{"type":"object","required":["x","y"],"properties":{"id":{"title":"Credential Subject ID","type":"string","format":"uri"},"x":{"type":"string"},"y":{"type":"string"}}}}}`
	AuthBJJCredentialSchemaType    = "https://schema.iden3.io/core/jsonld/auth.jsonld#AuthBJJCredential"
)

// SchemaFormat type
type SchemaFormat string

const (
	// JSONLD JSON-LD schema format
	JSONLD SchemaFormat = "json-ld"

	// JSON JSON schema format
	JSON SchemaFormat = "json"
)

// SchemaAttrs is a collection of schema attributes
type SchemaAttrs []string

// String satisfies the Stringer interface for SchemaAttrs
func (a SchemaAttrs) String() string {
	if len(a) == 0 {
		return ""
	}
	return strings.Join(a, ", ")
}

// SchemaAttrsFromString is an SchemaAttrs constructor from a string with  comma separated attributes
func SchemaAttrsFromString(commaAttrs string) SchemaAttrs {
	attrs := strings.Split(commaAttrs, ",")
	schemaAttrs := make(SchemaAttrs, len(attrs))
	for i, attr := range attrs {
		w := strings.TrimSpace(attr)
		if w != "" {
			schemaAttrs[i] = w
		}
	}
	return schemaAttrs
}

// Schema defines a domain.Schema entity
type Schema struct {
	ID         uuid.UUID
	IssuerDID  core.DID
	URL        string
	Type       string
	Hash       core.SchemaHash
	Attributes SchemaAttrs
	CreatedAt  time.Time
}
