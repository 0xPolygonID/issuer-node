package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ConnectionsRepository defines the available methods for connections repository
type ConnectionsRepository interface {
	Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error)
	Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID w3c.DID) error
	DeleteCredentials(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID w3c.DID) error
	GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID w3c.DID) (*domain.Connection, error)
	GetByUserID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, userDID w3c.DID) (*domain.Connection, error)
	GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, query string) ([]*domain.Connection, error)
	GetAllWithCredentialsByIssuerID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, query string) ([]*domain.Connection, error)
	GetByUserSessionID(ctx context.Context, conn db.Querier, sessionID uuid.UUID) (*domain.Connection, error)
	SaveUserAuthentication(ctx context.Context, conn db.Querier, connID uuid.UUID, sessID uuid.UUID, mTime time.Time) error
}
