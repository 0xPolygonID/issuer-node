package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
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
	require.NoError(t, err)
	return id
}

// CreateSchema creates an entry in schema table
func (f *Fixture) CreateSchema(t *testing.T, ctx context.Context, s *domain.Schema) {
	t.Helper()
	require.NoError(t, f.schemaRepository.Save(ctx, s))
}

// GetDefaultAuthClaimOfIssuer returns the default auth claim of an issuer just created
func (f *Fixture) GetDefaultAuthClaimOfIssuer(t *testing.T, issuerID string) *domain.Claim {
	t.Helper()
	ctx := context.Background()
	did, err := w3c.ParseDID(issuerID)
	assert.NoError(t, err)
	claims, _, err := f.claimRepository.GetAllByIssuerID(ctx, f.storage.Pgx, *did, &ports.ClaimsFilter{})
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
		ID:           fmt.Sprintf("http://localhost/api/v2/credentials/%s", claimID),
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
			ID:              fmt.Sprintf("http://localhost/v2/%s/credentials/revocation/status/%d", identity, revNonce),
			Type:            "SparseMerkleTreeProof",
			RevocationNonce: uint64(revNonce),
		},
		Issuer: identity,
		CredentialSchema: verifiable.CredentialSchema{
			ID:   "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type: "JsonSchemaValidator2018",
		},
		RefreshService: &verifiable.RefreshService{
			ID:   "https://refresh-service.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	assert.NoError(t, err)

	err = claim.Data.Set(vc)
	assert.NoError(t, err)

	return claim
}

// NewClaimWithEncryptionKey fixture
func (f *Fixture) NewClaimWithEncryptionKey(t *testing.T, identity string) *domain.Claim {
	t.Helper()
	const n = 32
	claimID, err := uuid.NewUUID()
	assert.NoError(t, err)

	nonce := int64(n)
	revNonce := domain.RevNonceUint64(nonce)
	claim := &domain.Claim{
		ID:              claimID,
		Identifier:      &identity,
		Issuer:          identity,
		SchemaHash:      "ca938857241db9451ea329256b9c06e8",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "KYCAgeCredential",
		OtherIdentifier: "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		Expiration:      0,
		Version:         0,
		RevNonce:        revNonce,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		EncryptedData:   common.ToPointer("eyJjaXBoZXJ0ZXh0IjoiZF9yMnl1dmZsR1NfLXhKMGNYM1h4Tk10MV9zb0tLUXRKWXBZYUR0a3UyY21ManFpbk54WENlNWo3X1lDLW96T05INzRmblRlc2dpTFpxcnYyajdfVXZ4czkyZ1F6QzdSRXBUaTJ3dEFta3FaNHBYdFQ5amxtYmQwVVFKRlFvc2xpRThkb0w2WlhRTHZwSGFTZ29USEVjNHdTRzVtUW8wdThXS0lrUWYxWm9NUEZHMjlDODAxNXBrZlZaMXdnRGRTeXNHdExaaGhYSXhSYWJtaWE5NmpmVXM0LWQ2RWtOX0Z6MC1Wc0xyMjZWa2Zrb1lweGpZUHRjTjA4UWI2TlBpdDhfcnNXa1ZzZXd4LW9sSE5BTVZ4b2I4QzRncjFWRnluT2pXZVdocXFocTg2bFc4eDlLQjdOR3hoWmdmV2VGeE9CU0k1U0hFUlRGb2VkRUdJVGhLSzlOMW10UklwRGRmQXRmWThfbGRtaWJjZHZLYS1sVm1yVjAwMkdKZWZsTmVmUE9PbEpFVnJmM3JQRW9rd2VNUG1mMzV3OXMxR2pRaFg1OW5tODA5dmhYSVRuSm01QXFWbTBBTGZsZWQzZVVCOXhpYVJ6N0ozcW5vUWdVUVQ0X0lSbkZPU25rNVBQZGpoUDA3WTgzXzZkQm1QdXZsRDBrYzVyMThBclY4UXhYc3NpTXVNWk1HR0ZyS2l5N0hwSzhnX1ZjZ3Z1R2FJY1BxUjg4UGI4Z254ZExWX3h5RDdwV1RsM20zdzNMZEVmQXlvX1VIRHVxR2dJTnB6UFVfb01sbkhMUVpyaWk1OFJPeE4yZFJZamZka196Wmw3TUFTSmhNeGFPNkZGck1uOHlhQnQwRmt6Rks2YnRocmFPRGFmQ00xVDVzb3JEU0tQeDRJTzRpNVltZjRHdnd1WTFtY3hmX29IMk1ZLWNMSFcwV3RTYnJ0WFBKUEVPUjN6WFVaTmxQLXdrd2VRaGw1YXN4YjZIOWhlQjJaMUthT1NwWWM1UGRVdzItZi1EU011S1ZzbVRUTFZaYTZWV1BvNEJJdTlkM0w3cUd4VVVoREYxaGtGRVoxbHhqdnU5Uk83WldnVzdDZHJsUERtWEpJZjA5ODk1ang4bThvSWhWbHlkaG1seU8tZzFBazMzb2h1TmhFc0NWR1JwYXhiNkpSWUpESEZFV0VMeHRRYUIxOW5UeW9IYlBXeTA5blFNbUZDektEMUJRLW5YS2Z0alltM3ZObkVZMldtWDNxNnZnTnEzLWIzTkFKLXNuNlVXdW1fQ0lTcG1zUFRDOWZOWXBsN3NOdDJ3Uk9qSTZVZlY3Q1YtUWZ3WkgxQnRueFozQ0ZTUTNPWHc5bEllRWxyN09LZkhTQVB6bEZialpaRm93OUNUWkI0djJpcWRpZ2N1UHVsWUptTzU2VDRJZ0tETTRLZ2l4WFo3UFZ5aGR1ZW02Vk1SRmd6Z1Q2UWFzVS00QXdSbGJmVzFNRHJZNGUxRElZVlJpdk5FbVVvSEd5cFdUU1NBcElkd3RrZTVfWVJpX2pFLXFFbFhGZFl1UERfaVhJdGU5aHVXdlE3RkdSY202TVpfbzl2cjE3ZGJyOXI1X0VUM3dEalhVcV9DZVJHNHhseVNKd1d2Y04zRWE4aUNVa09rVE0zOFNJMERoNFdrSUp1VktQTUVFRzF0dDQzWFkiLCJlbmNyeXB0ZWRfa2V5IjoiQU9DWl85czhyam9vQ2xCNEhjdGJTZzhpQUwydUVtZjJKNEdOLUhKSVFMaEZtcGI2cHlSSWt3IiwiaGVhZGVyIjp7ImFsZyI6IkVDREgtRVMrQTI1NktXIiwiZXBrIjp7ImNydiI6IlAtMjU2Iiwia3R5IjoiRUMiLCJ4IjoiaEJTZ1NWNEt3U3hUYVQtN0dodVBvYUpxWDktQmVkekFmU1RUekhiQzZsRSIsInkiOiJxZWtWZG5XTlZzb1R0Y3U1Y0RMdTd5T1NEZzFmSENVU20tTldNS1k1VnFvIn0sImtpZCI6InR1LWtpZCJ9LCJpdiI6IjBERG94OUlsZUItSDdEZGgiLCJwcm90ZWN0ZWQiOiJleUpoYkdjaU9pSkZRMFJJTFVWVEswRXlOVFpMVnlJc0ltVnVZeUk2SWtFeU5UWkhRMDBpTENKbGNHc2lPbnNpWTNKMklqb2lVQzB5TlRZaUxDSnJkSGtpT2lKRlF5SXNJbmdpT2lKb1FsTm5VMVkwUzNkVGVGUmhWQzAzUjJoMVVHOWhTbkZZT1MxQ1pXUjZRV1pUVkZSNlNHSkRObXhGSWl3aWVTSTZJbkZsYTFaa2JsZE9Wbk52VkhSamRUVmpSRXgxTjNsUFUwUm5NV1pJUTFWVGJTMU9WMDFMV1RWV2NXOGlmU3dpYTJsa0lqb2lkSFV0YTJsa0lpd2lkSGx3SWpvaVlYQndiR2xqWVhScGIyNHZhV1JsYmpOamIyMXRMV1Z1WTNKNWNIUmxaQzFxYzI5dUluMCIsInRhZyI6ImFCaTFIVldhMEpURW9DYlNfYnRHSEEifQ"),
		ContextUrl:      common.ToPointer("https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"),
	}

	err = claim.CredentialStatus.Set(verifiable.CredentialStatus{
		ID:              fmt.Sprintf("http://localhost/v2/%s/credentials/revocation/status/%d", identity, revNonce),
		Type:            "SparseMerkleTreeProof",
		RevocationNonce: uint64(revNonce),
	})
	assert.NoError(t, err)
	return claim
}
