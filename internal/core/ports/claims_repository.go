package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ClaimsRepository is the interface that defines the available methods
type ClaimsRepository interface {
	Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error)
	GetRevoked(ctx context.Context, conn db.Querier, currentState string) ([]*domain.Claim, error)
	Revoke(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error
	RevokeNonce(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error
	GetByRevocationNonce(ctx context.Context, conn db.Querier, identifier *w3c.DID, revocationNonce domain.RevNonceUint64) ([]*domain.Claim, error)
	GetByIdAndIssuer(ctx context.Context, conn db.Querier, identifier *w3c.DID, claimID uuid.UUID) (*domain.Claim, error)
	FindOneClaimBySchemaHash(ctx context.Context, conn db.Querier, subject *w3c.DID, schemaHash string) (*domain.Claim, error)
	GetAllByIssuerID(ctx context.Context, conn db.Querier, identifier w3c.DID, filter *ClaimsFilter) ([]*domain.Claim, uint, error)
	GetNonRevokedByConnectionAndIssuerID(ctx context.Context, conn db.Querier, connID uuid.UUID, issuerID w3c.DID) ([]*domain.Claim, error)
	GetAllByState(ctx context.Context, conn db.Querier, did *w3c.DID, state *merkletree.Hash) (claims []domain.Claim, err error)
	GetAllByStateWithMTProof(ctx context.Context, conn db.Querier, did *w3c.DID, state *merkletree.Hash) (claims []domain.Claim, err error)
	UpdateState(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error)
	GetAuthClaimsForPublishing(ctx context.Context, conn db.Querier, identifier *w3c.DID, publishingState string, schemaHash string) ([]*domain.Claim, error)
	UpdateClaimMTP(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error
	GetClaimsIssuedForUser(ctx context.Context, conn db.Querier, identifier w3c.DID, userDID w3c.DID, linkID uuid.UUID) ([]*domain.Claim, error)
	GetByStateIDWithMTPProof(ctx context.Context, conn db.Querier, did *w3c.DID, state string) (claims []*domain.Claim, err error)
}
