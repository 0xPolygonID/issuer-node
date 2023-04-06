package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IdentityStateRepository interface that defines the available methods
type IdentityStateRepository interface {
	Save(ctx context.Context, conn db.Querier, state domain.IdentityState) error
	GetLatestStateByIdentifier(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.IdentityState, error)
	GetStatesByStatus(ctx context.Context, conn db.Querier, status domain.IdentityStatus) ([]domain.IdentityState, error)
	GetStates(ctx context.Context, conn db.Querier, issuerDID core.DID) ([]domain.IdentityState, error)
	GetStatesByStatusAndIssuerID(ctx context.Context, conn db.Querier, status domain.IdentityStatus, issuerID core.DID) ([]domain.IdentityState, error)
	UpdateState(ctx context.Context, conn db.Querier, state *domain.IdentityState) (int64, error)
}
