package schema

import (
	"context"
	"encoding/json"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/loaders"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// Process constructs a core.Claim entity from a verifiable credential
// TODO: Test it
func Process(ctx context.Context, schemaURL, credentialType string, credential verifiable.W3CCredential, options *processor.CoreClaimOptions) (*core.Claim, error) {
	var parser processor.Parser
	var validator processor.Validator

	pr := &processor.Processor{}
	validator = jsonSuite.Validator{}
	parser = jsonSuite.Parser{}

	pr = processor.InitProcessorOptions(pr, processor.WithValidator(validator), processor.WithParser(parser), processor.WithSchemaLoader(getLoader(schemaURL)))

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

// LoadSchema loads schema from url
func LoadSchema(ctx context.Context, url string) (jsonSuite.Schema, error) {
	schemaBytes, _, err := load(ctx, url)
	if err != nil {
		return jsonSuite.Schema{}, err
	}

	var schema jsonSuite.Schema
	err = json.Unmarshal(schemaBytes, &schema)

	return schema, err
}

// getLoader returns corresponding loader (according to url schema)
// By now, only http
func getLoader(url string) processor.SchemaLoader {
	return &loaders.HTTP{URL: url}
}

// load returns schema content by url
func load(ctx context.Context, schemaURL string) (schema []byte, extension string, err error) {
	var schemaBytes []byte
	schemaBytes, _, err = getLoader(schemaURL).Load(ctx)
	if err != nil {
		return nil, "", err
	}

	return schemaBytes, string(domain.JSONLD), nil
}
