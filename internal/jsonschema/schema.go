package jsonschema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/utils"
	"github.com/mitchellh/mapstructure"

	"github.com/polygonid/sh-id-platform/internal/loader"
)

// Attributes is a list of Attribute entities
type Attributes []Attribute

// SchemaAttrs converts jsonschema.Attributes into []string]
func (a Attributes) SchemaAttrs() []string {
	out := make([]string, len(a))
	for i, attr := range a {
		out[i] = attr.String()
	}
	return out
}

// Attribute represents a json schema attribute
type Attribute struct {
	ID     string
	Title  string
	Type   string
	Format string
}

func (a Attribute) String() string {
	if a.Title != "" {
		return fmt.Sprintf("%s(%s)", a.Title, a.ID)
	}
	return a.ID
}

// JSONSchema provides some methods to load a schema and do some inspections over it.
type JSONSchema struct {
	content map[string]any
}

// Load loads the json file doing some validations..
func Load(ctx context.Context, loader loader.Loader) (*JSONSchema, error) {
	pr := processor.InitProcessorOptions(
		&processor.Processor{},
		processor.WithValidator(jsonSuite.Validator{}),
		processor.WithParser(jsonSuite.Parser{}),
		processor.WithSchemaLoader(loader))
	raw, _, err := pr.Load(ctx)
	if err != nil {
		return nil, err
	}

	schema := &JSONSchema{content: make(map[string]any)}
	if err := json.Unmarshal(raw, &schema.content); err != nil {
		return nil, err
	}
	return schema, nil
}

// Attributes returns a list with the attributes in properties.credentialSubject.properties
func (s *JSONSchema) Attributes() (Attributes, error) {
	var props map[string]any
	var ok bool
	props, ok = s.content["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties field")
	}
	credSubject, ok := props["credentialSubject"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties.credentialSubject field")
	}
	props, ok = credSubject["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties.credentialSubject.properties field")
	}
	attrs := make([]Attribute, 0, len(props))
	for id, prop := range props {
		attr := Attribute{}
		if err := mapstructure.Decode(prop, &attr); err != nil {
			return nil, fmt.Errorf("parsing attribute <%s>: %w", prop, err)
		}
		attr.ID = id
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

// AttributeByID returns the attribute with this id or an error if not found
func (s *JSONSchema) AttributeByID(id string) (*Attribute, error) {
	attrs, err := s.Attributes()
	if err != nil {
		return nil, err
	}
	i := findIndexForSchemaAttribute(attrs, id)
	if i < 0 {
		return nil, fmt.Errorf("schema attribute <%s> not found", id)
	}
	return &attrs[i], nil
}

// JSONLdContext returns the value of $metadata.uris.jsonLdContext
func (s *JSONSchema) JSONLdContext() (string, error) {
	var metadata map[string]any
	var ok bool
	metadata, ok = s.content["$metadata"].(map[string]any)
	if !ok {
		return "", errors.New("missing $metadata field")
	}
	uris, ok := metadata["uris"].(map[string]any)
	if !ok {
		return "", errors.New("missing $metadata.uris field")
	}
	jsonLdContext, ok := uris["jsonLdContext"].(string)
	if !ok {
		return "", errors.New("missing $metadata.uris.jsonLdContext field")
	}
	return jsonLdContext, nil
}

// SchemaHash calculates the hash of a schemaType
func (s *JSONSchema) SchemaHash(schemaType string) (core.SchemaHash, error) {
	jsonLdContext, err := s.JSONLdContext()
	if err != nil {
		return core.SchemaHash{}, err
	}
	id := jsonLdContext + "#" + schemaType
	return utils.CreateSchemaHash([]byte(id)), nil
}

func findIndexForSchemaAttribute(attributes Attributes, name string) int {
	for i, attribute := range attributes {
		if attribute.ID == name {
			return i
		}
	}
	return -1
}
