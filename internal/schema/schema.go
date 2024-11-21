package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core/v2"
	jsonSuiteV1 "github.com/iden3/go-schema-processor/json"
	jsonSuite "github.com/iden3/go-schema-processor/v2/json"
	"github.com/iden3/go-schema-processor/v2/processor"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
)

var (
	ErrLoadSchema   = errors.New("cannot load schema")          // ErrLoadSchema Cannot process schema
	ErrValidateData = errors.New("error validating claim data") // ErrValidateData Cannot process schema
	ErrParseClaim   = errors.New("error parsing claim")         // ErrParseClaim Cannot process schema
)

// LoadSchema loads schema from url
// TODO: Refactor to avoid v1 dependencies
func LoadSchema(ctx context.Context, loader loader.DocumentLoader, jsonSchemaURL string) (jsonSuiteV1.Schema, error) {
	var schema jsonSuiteV1.Schema

	schemaBytes, err := jsonschema.Load(ctx, jsonSchemaURL, loader)
	if err != nil {
		return schema, err
	}
	err = json.Unmarshal(schemaBytes.BytesNoErr(), &schema)

	return schema, err
}

// FromClaimModelToW3CCredential JSON-LD response base on claim
func FromClaimModelToW3CCredential(claim domain.Claim) (*verifiable.W3CCredential, error) {
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

	var mtpProof *verifiable.Iden3SparseMerkleTreeProof

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

// Process data and schema and create Index and Value slots
func Process(ctx context.Context, loader loader.DocumentLoader, schemaURL string, credential verifiable.W3CCredential, options *processor.CoreClaimOptions) (*core.Claim, error) {
	var parser processor.Parser
	var validator processor.Validator

	pr := &processor.Processor{}
	validator = jsonSuite.Validator{}
	parser = jsonSuite.Parser{}

	pr = processor.InitProcessorOptions(
		pr,
		processor.WithDocumentLoader(loader),
		processor.WithValidator(validator),
		processor.WithParser(parser))

	schema, err := pr.Load(ctx, schemaURL)
	if err != nil {
		log.Error(ctx, "error loading schema", "err", err)
		return nil, ErrLoadSchema
	}

	jsonCredential, err := json.Marshal(credential)
	if err != nil {
		log.Error(ctx, "error marshalling credential", "err", err)
		return nil, err
	}

	err = pr.ValidateData(jsonCredential, schema)
	if err != nil {
		log.Error(ctx, "error validating claim data", "err", err)
		return nil, ErrValidateData
	}

	claim, err := pr.ParseClaim(ctx, credential, options)
	if err != nil {
		log.Error(ctx, "error parsing claim", "err", err)
		return nil, NewParseClaimError(err.Error())
	}
	return claim, nil
}
