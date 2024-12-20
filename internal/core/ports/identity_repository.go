package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IdentityRepository is the interface implemented by the identity service
type IdentityRepository interface {
	Save(ctx context.Context, conn db.Querier, identity *domain.Identity) error
	GetByID(ctx context.Context, conn db.Querier, identifier w3c.DID) (*domain.Identity, error)
	Get(ctx context.Context, conn db.Querier) (identities []domain.IdentityDisplayName, err error)
	GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*w3c.DID, err error)
	HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error)
	HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error)
	UpdateDisplayName(ctx context.Context, conn db.Querier, identity *domain.Identity) error
}
