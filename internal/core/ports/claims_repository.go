package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type ClaimsRepository interface {
	Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error)
	FindOneClaimBySchemaHash(ctx context.Context, conn db.Querier, subject *core.DID, schemaHash string) (*domain.Claim, error)
}
