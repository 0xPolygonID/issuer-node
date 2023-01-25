package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

const (
	defualtAuthClaims = 1
)

// CreateClaim fixture
func (f *Fixture) CreateClaim(t *testing.T, claim *domain.Claim) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id, err := f.claimRepository.Save(ctx, f.storage.Pgx, claim)
	assert.NoError(t, err)
	return id
}

// GetDefaultAuthClaimOfIssuer returns the default auth claim of an issuer just created
func (f *Fixture) GetDefaultAuthClaimOfIssuer(t *testing.T, issuerID string) *domain.Claim {
	t.Helper()
	ctx := context.Background()
	did, err := core.ParseDID(issuerID)
	assert.NoError(t, err)
	claims, err := f.claimRepository.GetAllByIssuerID(ctx, f.storage.Pgx, did, &ports.Filter{})
	assert.NoError(t, err)
	require.Equal(t, len(claims), defualtAuthClaims)

	return claims[0]
}

// NewClaim fixture
// nolint
func (f *Fixture) NewClaim(t *testing.T, identity string) *domain.Claim {
	t.Helper()

	claimID, err := uuid.NewUUID()
	assert.NoError(t, err)

	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)
	claim := &domain.Claim{
		ID:              claimID,
		Identifier:      &identity,
		Issuer:          identity,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		Expiration:      0,
		Version:         0,
		RevNonce:        revNonce,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	}

	vc := verifiable.W3CCredential{
		ID:           fmt.Sprintf("http://localhost/api/v1/claim/%s", claimID),
		Context:      []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
		Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
		IssuanceDate: common.ToPointer(time.Now().UTC()),
		CredentialSubject: map[string]interface{}{
			"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
			"birthday":     19960424,
			"documentType": 2,
			"type":         "KYCAgeCredential",
		},
		CredentialStatus: verifiable.CredentialStatus{
			ID:              fmt.Sprintf("http://localhost/api/v1/identities/%s/claims/revocation/status/%d", identity, revNonce),
			Type:            "SparseMerkleTreeProof",
			RevocationNonce: uint64(revNonce),
		},
		Issuer: identity,
		CredentialSchema: verifiable.CredentialSchema{
			ID:   "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type: "JsonSchemaValidator2018",
		},
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	assert.NoError(t, err)

	err = claim.Data.Set(vc)
	assert.NoError(t, err)

	return claim
}
