package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"
)

// Publisher - Define the interface for publisher services
type Publisher interface {
	PublishState(ctx context.Context, identity *core.DID) (*string, error)
	CheckTransactionStatus(ctx context.Context)
}
