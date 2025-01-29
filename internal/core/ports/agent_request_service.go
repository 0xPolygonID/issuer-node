package ports

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
)

// NewAgentRequest validates the inputs and returns a new AgentRequest
func NewAgentRequest(basicMessage *comm.BasicMessage) (*AgentRequest, error) {
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

	ID, err := uuid.Parse(basicMessage.ID)
	if err != nil {
		return nil, err
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field cannot be empty")
	}

	return &AgentRequest{
		Body:      basicMessage.Body,
		UserDID:   fromDID,
		IssuerDID: toDID,
		ThreadID:  basicMessage.ThreadID,
		Typ:       basicMessage.Typ,
		Type:      basicMessage.Type,
		ID:        ID,
	}, nil
}
