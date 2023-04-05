package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// LinkRepository the interface that defines the available methods
type LinkRepository interface {
	Save(ctx context.Context, conn db.Querier, link *domain.Link) (*uuid.UUID, error)
	GetByID(ctx context.Context, issuerID core.DID, id uuid.UUID) (*domain.Link, error)
	GetAll(ctx context.Context, issuerDID core.DID, status LinkStatus, query *string) ([]domain.Link, error)
	Delete(ctx context.Context, id uuid.UUID, issuerDID core.DID) error
}
