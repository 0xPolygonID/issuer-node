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
	"github.com/xeipuuv/gojsonschema"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
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
	ID         string
	Title      string
	Type       string
	Format     string
	Properties map[string]any `json:"-"`
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

// ValidateCredentialSubject checks if the given credential subject matches the schema definition
func (s *JSONSchema) ValidateCredentialSubject(cSubject domain.CredentialSubject) error {
	valJSONSchema := s.getValidationJSONSchema()
	err := valJSONSchema.formatForValidation()
	if err != nil {
		return err
	}

	return valJSONSchema.validateCredentialSubject(cSubject)
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
	attrs, err := processProperties(props)
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func (s *JSONSchema) properties() (map[string]any, error) {
	props, ok := s.content["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("missing properties field")
	}

	return props, nil
}

func (s *JSONSchema) setRequired(required []string) {
	s.content["required"] = required
}

func (s *JSONSchema) removeIDFromRequired() error {
	props, err := s.properties()
	if err != nil {
		return err
	}

	credentialSubject, ok := props["credentialSubject"].(map[string]any)
	if !ok {
		return errors.New("malformed credentialSubject")
	}

	credRequired, ok := credentialSubject["required"].([]any)
	if !ok {
		return errors.New("malformed required")
	}

	required := make([]string, 0)
	for _, credRequired := range credRequired {
		credRequiredStr, ok := credRequired.(string)
		if !ok {
			return fmt.Errorf("malformed required property inside credentialSubject %v", credRequired)
		}

		if credRequiredStr != "id" {
			required = append(required, credRequiredStr)
		}
	}

	credentialSubject["required"] = required
	props["credentialSubject"] = credentialSubject

	return nil
}

func (s *JSONSchema) formatForValidation() error {
	s.setRequired([]string{"credentialSubject"})

	return s.removeIDFromRequired()
}

func (s *JSONSchema) getValidationJSONSchema() JSONSchema {
	return JSONSchema{content: common.CopyMap(s.content)}
}

func (s *JSONSchema) validateCredentialSubject(cSubject domain.CredentialSubject) error {
	schemaBytes, err := json.Marshal(s.content)
	if err != nil {
		return errors.New("marshalling schema credential subject")
	}

	cSubjectBytes, err := json.Marshal(map[string]interface{}{"credentialSubject": cSubject})
	if err != nil {
		return errors.New("marshalling link credential subject")
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
	linkLoader := gojsonschema.NewStringLoader(string(cSubjectBytes))

	result, err := gojsonschema.Validate(schemaLoader, linkLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	var errStr string
	for _, e := range result.Errors() {
		errStr += fmt.Sprintf("%s \n", e.String())
	}

	return errors.New(errStr)
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

func processProperties(props map[string]any) ([]Attribute, error) {
	attrs := make([]Attribute, 0, len(props))
	for id, prop := range props {
		attr := Attribute{}
		if err := mapstructure.Decode(prop, &attr); err != nil {
			return nil, fmt.Errorf("parsing attribute <%s>: %w", prop, err)
		}
		attr.ID = id
		if len(attr.Properties) > 0 {
			attrs = append(attrs, Attribute{
				ID:   id,
				Type: "object",
			})
			attrs1, err := processProperties(attr.Properties)
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, attrs1...)
		} else {
			attrs = append(attrs, attr)
		}
	}
	return attrs, nil
}
