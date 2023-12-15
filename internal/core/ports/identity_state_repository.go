package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/db"
)

// IdentityStateRepository interface that defines the available methods
type IdentityStateRepository interface {
	Save(ctx context.Context, conn db.Querier, state domain.IdentityState) error
	GetLatestStateByIdentifier(ctx context.Context, conn db.Querier, identifier *w3c.DID) (*domain.IdentityState, error)
	GetStatesByStatus(ctx context.Context, conn db.Querier, status domain.IdentityStatus) ([]domain.IdentityState, error)
	GetStates(ctx context.Context, conn db.Querier, issuerDID w3c.DID) ([]domain.IdentityState, error)
	GetStatesByStatusAndIssuerID(ctx context.Context, conn db.Querier, status domain.IdentityStatus, issuerID w3c.DID) ([]domain.IdentityState, error)
	UpdateState(ctx context.Context, conn db.Querier, state *domain.IdentityState) (int64, error)
	GetGenesisState(ctx context.Context, conn db.Querier, identifier string) (*domain.IdentityState, error)
}
