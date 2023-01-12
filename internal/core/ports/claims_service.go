package ports

import (
	"context"
	"encoding/json"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// ClaimRequest struct
type ClaimRequest struct {
	Schema                jsonSuite.Schema
	DID                   *core.DID
	CredentialSchema      string
	CredentialSubject     json.RawMessage
	Expiration            *int64
	Type                  string
	Version               uint32
	SubjectPos            string
	MerklizedRootPosition string
}

// Validate ensures that a claim is correct
func (c *ClaimRequest) Validate() error {
	return nil
}

// NewClaimRequest returns a new claim object with the given parameters
func NewClaimRequest(schema jsonSuite.Schema, did *core.DID, credentialSchema string, credentialSubject json.RawMessage, expiration *int64, typ string, cVersion *uint32, subjectPos *string, merklizedRootPosition *string) *ClaimRequest {
	var version uint32
	var subject, merklizedRP string
	if cVersion != nil {
		version = *cVersion
	}
	if subjectPos != nil {
		subject = *subjectPos
	}
	if merklizedRootPosition != nil {
		merklizedRP = *merklizedRootPosition
	}

	return &ClaimRequest{
		Schema:                schema,
		DID:                   did,
		CredentialSchema:      credentialSchema,
		CredentialSubject:     credentialSubject,
		Expiration:            expiration,
		Type:                  typ,
		Version:               version,
		SubjectPos:            subject,
		MerklizedRootPosition: merklizedRP,
	}
}

// ClaimsService is the interface implemented by the claim service
type ClaimsService interface {
	CreateVC(ctx context.Context, claimReq *ClaimRequest, nonce uint64) (verifiable.W3CCredential, error)
	GetAuthClaim(ctx context.Context, did *core.DID) (*domain.Claim, error)
	SendClaimOfferPushNotification(ctx context.Context, claim *domain.Claim) error
	GetRevocationSource(issuerDID string, nonce uint64) interface{}
	Save(ctx context.Context, claim *domain.Claim) (*domain.Claim, error)
	Revoke(ctx context.Context, id string, nonce uint64, description string) error
}
