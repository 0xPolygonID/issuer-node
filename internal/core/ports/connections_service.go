package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// ConnectionsService  is the interface implemented by the Connections service
type ConnectionsService interface {
	Delete(ctx context.Context, id uuid.UUID) error
	GetByIDAndIssuerID(ctx context.Context, id uuid.UUID, issuerDID core.DID) (*domain.Connection, error)
}
