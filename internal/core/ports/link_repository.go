package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

type LinkRepository interface {
	Save(ctx context.Context, link *domain.Link) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Link, error)
}
