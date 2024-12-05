package ports

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

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

// CreatePaymentRequestReq is the request for PaymentService.CreatePaymentRequest
type CreatePaymentRequestReq struct {
	IssuerDID w3c.DID
	UserDID   w3c.DID
	OptionID  uuid.UUID
	Creds     []protocol.PaymentRequestInfoCredentials
}

// PaymentService is the interface implemented by the payment service
type PaymentService interface {
	CreatePaymentRequest(ctx context.Context, req *CreatePaymentRequestReq, baseURL string) (*protocol.PaymentRequestMessage, error)
	CreatePaymentRequestForProposalRequest(ctx context.Context, proposalRequest *protocol.CredentialsProposalRequestMessage) (*comm.BasicMessage, error)
	GetSettings() payments.Config
	VerifyPayment(ctx context.Context, paymentOptionID uuid.UUID, message *protocol.PaymentMessage) (bool, error)

	CreatePaymentOption(ctx context.Context, issuerDID *w3c.DID, name, description string, config *domain.PaymentOptionConfig) (uuid.UUID, error)
	GetPaymentOptions(ctx context.Context, issuerDID *w3c.DID) ([]domain.PaymentOption, error)
	GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error)
	DeletePaymentOption(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) error
}
