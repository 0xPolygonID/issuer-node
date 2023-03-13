package services

import (
	"context"
	"encoding/json"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/loader"
)

type schema struct {
	loaderFactory loader.Factory
}

// NewSchema creates a new schema service
func NewSchema(lf loader.Factory) ports.SchemaService {
	return &schema{
		loaderFactory: lf,
	}
}

// LoadSchema loads schema from url
func (s *schema) LoadSchema(ctx context.Context, url string) (jsonSuite.Schema, error) {
	schemaBytes, _, err := s.load(ctx, url)
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

	pr = processor.InitProcessorOptions(pr, processor.WithValidator(validator), processor.WithParser(parser), processor.WithSchemaLoader(s.loaderFactory(schemaURL)))

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

// FromClaimModelToW3CCredential JSON-LD response base on claim
func (s *schema) FromClaimModelToW3CCredential(claim domain.Claim) (*verifiable.W3CCredential, error) {
	var cred verifiable.W3CCredential

	err := json.Unmarshal(claim.Data.Bytes, &cred)
	if err != nil {
		return nil, err
	}
	if claim.CredentialStatus.Status == pgtype.Null {
		return nil, fmt.Errorf("credential status is not set")
	}

	proofs := make(verifiable.CredentialProofs, 0)

	var signatureProof *verifiable.BJJSignatureProof2021
	if claim.SignatureProof.Status != pgtype.Null {
		err = claim.SignatureProof.AssignTo(&signatureProof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, signatureProof)
	}

	var mtpProof *verifiable.Iden3SparseMerkleProof // nolint: staticcheck

	if claim.MTPProof.Status != pgtype.Null {
		err = claim.MTPProof.AssignTo(&mtpProof)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, mtpProof)

	}
	cred.Proof = proofs

	return &cred, nil
}

// load returns schema content by url
func (s *schema) load(ctx context.Context, schemaURL string) (schema []byte, extension string, err error) {
	var schemaBytes []byte
	schemaBytes, _, err = s.loaderFactory(schemaURL).Load(ctx)
	if err != nil {
		return nil, "", err
	}

	return schemaBytes, string(domain.JSONLD), nil
}
