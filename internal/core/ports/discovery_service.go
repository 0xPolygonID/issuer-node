package ports

import (
	"context"
	"fmt"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"
)

// DiscoveryService is the interface implemented by the discovery service
type DiscoveryService interface {
	Agent(ctx context.Context, req *AgentRequest) (*iden3comm.BasicMessage, error)
}

// NewDiscoveryAgentRequest validates the inputs and returns a new AgentRequest
func NewDiscoveryAgentRequest(basicMessage *comm.BasicMessage) (*AgentRequest, error) {
	var toDID, fromDID *w3c.DID
	var err error
	if basicMessage.To != "" {
		toDID, err = w3c.ParseDID(basicMessage.To)
		if err != nil {
			return nil, err
		}
	}

	if basicMessage.From != "" {
		fromDID, err = w3c.ParseDID(basicMessage.From)
		if err != nil {
			return nil, err
		}
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field cannot be empty")
	}

	if basicMessage.Type != protocol.DiscoverFeatureQueriesMessageType {
		return nil, fmt.Errorf("invalid type")
	}

	return &AgentRequest{
		Body:      basicMessage.Body,
		UserDID:   fromDID,
		IssuerDID: toDID,
		ThreadID:  basicMessage.ThreadID,
		Typ:       basicMessage.Typ,
		Type:      basicMessage.Type,
	}, nil
}
