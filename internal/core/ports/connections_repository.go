package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ConnectionsRepository defines the available methods for connections repository
type ConnectionsRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID core.DID) error
	DeleteCredentials(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID core.DID) error
	GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID core.DID) (*domain.Connection, error)
	GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerDID core.DID, query string) ([]*domain.Connection, error)
	GetAllWithCredentialsByIssuerID(ctx context.Context, conn db.Querier, issuerDID core.DID, query string) ([]*domain.Connection, error)
}
