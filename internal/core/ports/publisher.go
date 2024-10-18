package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// Publisher - Define the interface for publisher services
type Publisher interface {
	PublishState(ctx context.Context, identity *w3c.DID) (*domain.PublishedState, error)
	RetryPublishState(ctx context.Context, identifier *w3c.DID) (*domain.PublishedState, error)
	CheckTransactionStatus(ctx context.Context, identity *domain.Identity)
}
