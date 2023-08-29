package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

// IdentityService is the interface implemented by the identity service
type IdentityService interface {
	GetByDID(ctx context.Context, identifier core.DID) (*domain.Identity, error)
	Create(ctx context.Context, DIDMethod string, Blockchain, NetworkID, hostURL string) (*domain.Identity, error)
	SignClaimEntry(ctx context.Context, authClaim *domain.Claim, claimEntry *core.Claim) (*verifiable.BJJSignatureProof2021, error)
	Get(ctx context.Context) (identities []string, err error)
	UpdateState(ctx context.Context, did core.DID) (*domain.IdentityState, error)
	Exists(ctx context.Context, identifier core.DID) (bool, error)
	GetLatestStateByID(ctx context.Context, identifier core.DID) (*domain.IdentityState, error)
	GetKeyIDFromAuthClaim(ctx context.Context, authClaim *domain.Claim) (kms.KeyID, error)
	GetUnprocessedIssuersIDs(ctx context.Context) ([]*core.DID, error)
	HasUnprocessedStatesByID(ctx context.Context, identifier core.DID) (bool, error)
	HasUnprocessedAndFailedStatesByID(ctx context.Context, identifier core.DID) (bool, error)
	GetNonTransactedStates(ctx context.Context) ([]domain.IdentityState, error)
	UpdateIdentityState(ctx context.Context, state *domain.IdentityState) error
	GetTransactedStates(ctx context.Context) ([]domain.IdentityState, error)
	GetStates(ctx context.Context, issuerDID core.DID) ([]domain.IdentityState, error)
	CreateAuthenticationQRCode(ctx context.Context, serverURL string, issuerDID core.DID) (string, error)
	Authenticate(ctx context.Context, message string, sessionID uuid.UUID, serverURL string, issuerDID core.DID) (*protocol.AuthorizationResponseMessage, error)
	GetFailedState(ctx context.Context, identifier core.DID) (*domain.IdentityState, error)
}
