package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

// IndentityService is the interface implemented by the identity service
type IndentityService interface {
	Create(ctx context.Context, hostURL string) (*domain.Identity, error)
	SignClaimEntry(ctx context.Context, authClaim *domain.Claim, claimEntry *core.Claim) (*verifiable.BJJSignatureProof2021, error)
	Get(ctx context.Context) (identities []string, err error)
	UpdateState(ctx context.Context, did *core.DID) (*domain.IdentityState, error)
	GetUnprocessedIssuersIDs(ctx context.Context) ([]*core.DID, error)
	GetNonTransactedStates(ctx context.Context) ([]domain.IdentityState, error)
	GetLatestStateByID(ctx context.Context, identifier *core.DID) (*domain.IdentityState, error)
	GetKeyIDFromAuthClaim(ctx context.Context, authClaim *domain.Claim) (kms.KeyID, error)
	UpdateIdentityState(ctx context.Context, state *domain.IdentityState) error
}
