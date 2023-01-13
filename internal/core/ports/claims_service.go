package ports

import (
	"context"
	"encoding/json"
	"time"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// ClaimRequest struct
type ClaimRequest struct {
	Schema                string
	DID                   *core.DID
	CredentialSchema      string
	CredentialSubject     json.RawMessage
	Expiration            *time.Time
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
func NewClaimRequest(schema string, did *core.DID, credentialSchema string, credentialSubject json.RawMessage, expiration *int64, typ string, cVersion *uint32, subjectPos *string, merklizedRootPosition *string) *ClaimRequest {
	req := &ClaimRequest{
		Schema:            schema,
		DID:               did,
		CredentialSchema:  credentialSchema,
		CredentialSubject: credentialSubject,
		Type:              typ,
	}
	if expiration != nil {
		t := time.Unix(*expiration, 0)
		req.Expiration = &t
	}
	if cVersion != nil {
		req.Version = *cVersion
	}
	if subjectPos != nil {
		req.SubjectPos = *subjectPos
	}
	if merklizedRootPosition != nil {
		req.MerklizedRootPosition = *merklizedRootPosition
	}
	return req
}

// ClaimsService is the interface implemented by the claim service
type ClaimsService interface {
	CreateClaim(ctx context.Context, claimReq *ClaimRequest) (*domain.Claim, error)
	Revoke(ctx context.Context, id string, nonce uint64, description string) error
}
