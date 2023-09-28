package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// RevocationRepository interface that defines the available methods
type RevocationRepository interface {
	UpdateStatus(ctx context.Context, conn db.Querier, did *w3c.DID) ([]*domain.Revocation, error)
}
