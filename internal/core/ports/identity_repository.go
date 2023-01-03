package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// IndentityRepository is the interface implemented by the identity service
type IndentityRepository interface {
	Save(ctx context.Context, conn db.Querier, identity *domain.Identity) error
	GetByID(ctx context.Context, conn db.Querier, identifier *core.DID) (*domain.Identity, error)
}
