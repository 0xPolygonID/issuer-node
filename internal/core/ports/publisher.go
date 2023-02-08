package ports

import (
	"context"
	core "github.com/iden3/go-iden3-core"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// Publisher - Define the interface for publisher services
type Publisher interface {
	PublishState(ctx context.Context, identity *core.DID) (*domain.PublishedState, error)
	CheckTransactionStatus(ctx context.Context)
}
