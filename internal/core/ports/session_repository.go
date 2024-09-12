package ports

import (
	"context"

	"github.com/iden3/iden3comm/v2/protocol"
)

// SessionRepository defines the interface for managing sessions
type SessionRepository interface {
	Get(ctx context.Context, key string) (protocol.AuthorizationRequestMessage, error)
	Set(ctx context.Context, key string, value protocol.AuthorizationRequestMessage) error
}
