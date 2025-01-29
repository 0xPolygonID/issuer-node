package ports

import (
	"context"

	"github.com/iden3/iden3comm/v2"
)

// DiscoveryService is the interface implemented by the discovery service
type DiscoveryService interface {
	Agent(ctx context.Context, req *AgentRequest) (*iden3comm.BasicMessage, error)
}
