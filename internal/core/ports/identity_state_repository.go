package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type IdentityStateRepository interface {
	Save(ctx context.Context, conn db.Querier, state domain.IdentityState) error
}
