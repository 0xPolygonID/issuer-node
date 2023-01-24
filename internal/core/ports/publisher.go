package ports

import "context"

// Publisher - Define the interface for publisher services
type Publisher interface {
	PublishState(ctx context.Context)
	CheckTransactionStatus(ctx context.Context)
}
