package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ClaimsRepository is the interface that defines the available methods
type ClaimsRepository interface {
	Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error)
	Revoke(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error
	RevokeNonce(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error
	GetByRevocationNonce(ctx context.Context, conn db.Querier, identifier *core.DID, revocationNonce domain.RevNonceUint64) (*domain.Claim, error)
	GetByIdAndIssuer(ctx context.Context, conn db.Querier, identifier *core.DID, claimID uuid.UUID) (*domain.Claim, error)
	FindOneClaimBySchemaHash(ctx context.Context, conn db.Querier, subject *core.DID, schemaHash string) (*domain.Claim, error)
	GetAllByIssuerID(ctx context.Context, conn db.Querier, identifier core.DID, filter *ClaimsFilter) ([]*domain.Claim, error)
	GetNonRevokedByConnectionAndIssuerID(ctx context.Context, conn db.Querier, connID uuid.UUID, issuerID core.DID) ([]*domain.Claim, error)
	GetAllByState(ctx context.Context, conn db.Querier, did *core.DID, state *merkletree.Hash) (claims []domain.Claim, err error)
	GetAllByStateWithMTProof(ctx context.Context, conn db.Querier, did *core.DID, state *merkletree.Hash) (claims []domain.Claim, err error)
	UpdateState(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error)
	GetAuthClaimsForPublishing(ctx context.Context, conn db.Querier, identifier *core.DID, publishingState string, schemaHash string) ([]*domain.Claim, error)
	UpdateClaimMTP(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error
	GetClaimsIssuedForUser(ctx context.Context, conn db.Querier, identifier core.DID, userDID core.DID, linkID uuid.UUID) ([]*domain.Claim, error)
	GetByStateIDWithMTPProof(ctx context.Context, conn db.Querier, did *core.DID, state string) (claims []*domain.Claim, err error)
}
