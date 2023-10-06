package ports

import (
	"context"

	// core "github.com/iden3/go-iden3-core"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IndentityRepository is the interface implemented by the identity service
type RequestRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Request) error
	GetByID(ctx context.Context, conn db.Querier, id uuid.UUID)(domain.Responce,error)
	Get(ctx context.Context, conn db.Querier)([]*domain.Responce,error)
	UpdateStatus(ctx context.Context, conn db.Querier,id uuid.UUID)(int64, error)
	// GetByID(ctx context.Context, conn db.Querier, identifier core.DID) (*domain.Identity, error)
	// Get(ctx context.Context, conn db.Querier) (identities []string, err error)
	// GetUnprocessedIssuersIDs(ctx context.Context, conn db.Querier) (issuersIDs []*core.DID, err error)
	// HasUnprocessedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
	// HasUnprocessedAndFailedStatesByID(ctx context.Context, conn db.Querier, identifier *core.DID) (bool, error)
}
