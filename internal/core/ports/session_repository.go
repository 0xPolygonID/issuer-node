package ports

import (
	"context"

	"github.com/iden3/iden3comm/v2/protocol"

	link_state "github.com/polygonid/issuer-node/pkg/link"
)

// SessionRepository defines the interface for managing sessions
type SessionRepository interface {
	Get(ctx context.Context, key string) (protocol.AuthorizationRequestMessage, error)
	Set(ctx context.Context, key string, value protocol.AuthorizationRequestMessage) error
	SetLink(ctx context.Context, key string, value link_state.State) error
	GetLink(ctx context.Context, key string) (link_state.State, error)
}
