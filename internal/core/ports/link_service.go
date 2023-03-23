package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// LinkService - the interface that defines the available methods
type LinkService interface {
	Save(ctx context.Context, link domain.Link) (*uuid.UUID, error)
}
