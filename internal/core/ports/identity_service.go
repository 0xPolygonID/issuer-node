package ports

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

type IndentityService interface {
	Create(ctx context.Context, hostURL string) (*domain.Identity, error)
}
