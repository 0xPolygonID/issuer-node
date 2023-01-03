package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// IndentityService is the interface implemented by the identity service
type IndentityService interface {
	Create(ctx context.Context, hostURL string) (*domain.Identity, error)
}
