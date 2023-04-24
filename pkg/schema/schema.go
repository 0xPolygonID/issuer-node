package schema

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/loader"
)

var (
	ErrLoadSchema   = errors.New("cannot load schema")          // ErrProcessSchema Cannot process schema
	ErrValidateData = errors.New("error validating claim data") // ErrValidateData Cannot process schema
	ErrParseClaim   = errors.New("error parsing claim")         // ErrParseClaim Cannot process schema
)

// LoadSchema loads schema from url
func LoadSchema(ctx context.Context, loader loader.Loader) (jsonSuite.Schema, error) {
	var schema jsonSuite.Schema
	schemaBytes, _, err := loader.Load(ctx)
	if err != nil {
		return schema, err
	}
	err = json.Unmarshal(schemaBytes, &schema)

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

// FromClaimsModelToW3CCredential JSON-LD response base on claim
func FromClaimsModelToW3CCredential(credentials domain.Credentials) ([]*verifiable.W3CCredential, error) {
	w3Credentials := make([]*verifiable.W3CCredential, len(credentials))
	for i := range credentials {
		w3Cred, err := FromClaimModelToW3CCredential(*credentials[i])
		if err != nil {
			return nil, err
		}

		w3Credentials[i] = w3Cred
	}

	return w3Credentials, nil
}

// Process data and schema and create Index and Value slots
func Process(ctx context.Context, ld loader.Loader, credentialType string, credential verifiable.W3CCredential, options *processor.CoreClaimOptions) (*core.Claim, error) {
	var parser processor.Parser
	var validator processor.Validator
	pr := &processor.Processor{}

	validator = jsonSuite.Validator{}
	parser = jsonSuite.Parser{}

	pr = processor.InitProcessorOptions(pr, processor.WithValidator(validator), processor.WithParser(parser), processor.WithSchemaLoader(ld))

	schema, _, err := pr.Load(ctx)
	if err != nil {
		return nil, ErrLoadSchema
	}

	jsonCredential, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	err = pr.ValidateData(jsonCredential, schema)
	if err != nil {
		return nil, ErrValidateData
	}

	claim, err := pr.ParseClaim(ctx, credential, credentialType, schema, options)
	if err != nil {
		return nil, ErrParseClaim
	}
	return claim, nil
}
