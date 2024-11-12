package ports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/payments"
)

// Credential struct: TODO: This shouldn't be here with this name
type Credential struct {
	Type    string `json:"type"`
	Context string `json:"context"`
	// typss   protocol.CredentialIssuanceRequestMessage
}

// AgentProposalRequest struct
type AgentProposalRequest struct {
	Body        json.RawMessage
	ThreadID    string
	IssuerDID   *w3c.DID
	UserDID     *w3c.DID
	Typ         comm.MediaType
	Type        comm.ProtocolMessage
	Credentials []Credential
}

// NewAgentProposalRequest validates the inputs and returns a new AgentRequest
func NewAgentProposalRequest(basicMessage *comm.BasicMessage) (*protocol.CredentialsProposalRequestMessage, error) {
	if basicMessage.To == "" {
		return nil, fmt.Errorf("'to' field cannot be empty")
	}

	toDID, err := w3c.ParseDID(basicMessage.To)
	if err != nil {
		return nil, err
	}

	if basicMessage.From == "" {
		return nil, fmt.Errorf("'from' field cannot be empty")
	}

	fromDID, err := w3c.ParseDID(basicMessage.From)
	if err != nil {
		return nil, err
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field cannot be empty")
	}

	if basicMessage.Type != protocol.CredentialProposalRequestMessageType {
		return nil, fmt.Errorf("invalid type")
	}

	var body *protocol.CredentialsProposalRequestBody
	if err := json.Unmarshal(basicMessage.Body, body); err != nil {
		return nil, err
	}

	return &protocol.CredentialsProposalRequestMessage{
		ID:       basicMessage.ID,
		ThreadID: basicMessage.ThreadID,
		Typ:      basicMessage.Typ,
		Type:     basicMessage.Type,
		Body:     *body,
		To:       toDID.String(),
		From:     fromDID.String(),
	}, nil
}

// PaymentService is the interface implemented by the payment service
type PaymentService interface {
	CreatePaymentRequest(ctx context.Context, issuerDID *w3c.DID, senderDID *w3c.DID, signingKey string, credContext string, credType string) (*comm.BasicMessage, error)
	CreatePaymentRequestForProposalRequest(ctx context.Context, proposalRequest *protocol.CredentialsProposalRequestMessage) (*comm.BasicMessage, error)
	GetSettings() payments.Settings
	VerifyPayment(ctx context.Context, recipient common.Address, message comm.BasicMessage) (bool, error)
}
