package services

import (
	"context"
	"encoding/json"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/loaders"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type schema struct {
	storage *db.Storage
}

// NewSchema creates a new schema service
func NewSchema(storage *db.Storage) ports.SchemaService {
	return &schema{storage: storage}
}

// Load returns schema content by url
func (s *schema) Load(ctx context.Context, schemaURL string) (schema []byte, extension string, err error) {
	var schemaBytes []byte
	schemaBytes, _, err = s.GetLoader(schemaURL).Load(ctx)
	if err != nil {
		return nil, "", err
	}

	return schemaBytes, string(domain.JSONLD), nil
}

// LoadSchema loads schema from url
func (s *schema) LoadSchema(ctx context.Context, url string) (jsonSuite.Schema, error) {
	schemaBytes, _, err := s.Load(ctx, url)
	if err != nil {
		return jsonSuite.Schema{}, err
	}

	var schema jsonSuite.Schema
	err = json.Unmarshal(schemaBytes, &schema)

	return schema, err
}

// Process data and schema and create Index and Value slots
func (s *schema) Process(ctx context.Context, schemaURL, credentialType string, credential verifiable.W3CCredential, options *processor.CoreClaimOptions) (*core.Claim, error) {
	var parser processor.Parser
	var validator processor.Validator
	pr := &processor.Processor{}

	validator = jsonSuite.Validator{}
	parser = jsonSuite.Parser{}

	pr = processor.InitProcessorOptions(pr, processor.WithValidator(validator), processor.WithParser(parser), processor.WithSchemaLoader(s.GetLoader(schemaURL)))

	schema, _, err := pr.Load(ctx)
	if err != nil {
		return nil, err
	}

	jsonCredential, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	err = pr.ValidateData(jsonCredential, schema)
	if err != nil {
		return nil, err
	}

	claim, err := pr.ParseClaim(ctx, credential, credentialType, schema, options)
	if err != nil {
		return nil, err
	}
	return claim, nil
}

// GetLoader returns corresponding loader (according to url schema)
func (s *schema) GetLoader(_url string) processor.SchemaLoader {
	return &loaders.HTTP{URL: _url}
}
