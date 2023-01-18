package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// IdentityStateService - is the interface implemented by the identity state service
type IdentityStateService interface {
	UpdateIdentityClaims(ctx context.Context, did *core.DID) (*domain.IdentityState, error)
}
