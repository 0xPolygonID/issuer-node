package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateClaimRequest struct
type CreateClaimRequest struct {
	DID                   *core.DID
	Schema                string
	CredentialSubject     map[string]any
	Expiration            *time.Time
	Type                  string
	Version               uint32
	SubjectPos            string
	MerklizedRootPosition string
}

// NewCreateClaimRequest returns a new claim object with the given parameters
func NewCreateClaimRequest(did *core.DID, credentialSchema string, credentialSubject map[string]any, expiration *int64, typ string, cVersion *uint32, subjectPos *string, merklizedRootPosition *string) *CreateClaimRequest {
	req := &CreateClaimRequest{
		DID:               did,
		Schema:            credentialSchema,
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
	CreateClaim(ctx context.Context, claimReq *CreateClaimRequest) (*domain.Claim, error)
	Revoke(ctx context.Context, id string, nonce uint64, description string) error
	GetByID(ctx context.Context, issID *core.DID, id uuid.UUID) (*verifiable.W3CCredential, error)
}
