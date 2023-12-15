package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/db"
)

// IndentityRepository is the interface implemented by the identity service
type IndentityRepository interface {
	Save(ctx context.Context, conn db.Querier, identity *domain.Identity) error
	GetByID(ctx context.Context, conn db.Querier, identifier w3c.DID) (*domain.Identity, error)
	Get(ctx context.Context, conn db.Querier) (identities []string, err error)
	GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*w3c.DID, err error)
	HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error)
	HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *w3c.DID) (bool, error)
}
