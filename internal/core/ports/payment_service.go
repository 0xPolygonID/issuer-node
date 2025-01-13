package ports

import (
	"context"
	"math/big"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

// BlockchainPaymentStatus represents the status of a payment
type BlockchainPaymentStatus int

const (
	// BlockchainPaymentStatusPending - Payment is still pending
	BlockchainPaymentStatusPending BlockchainPaymentStatus = iota
	// BlockchainPaymentStatusCancelled - Payment was cancelled
	BlockchainPaymentStatusCancelled
	// BlockchainPaymentStatusSuccess - Payment was successful
	BlockchainPaymentStatusSuccess
	// BlockchainPaymentStatusFailed - Payment has failed
	BlockchainPaymentStatusFailed
	// BlockchainPaymentStatusUnknown - Payment status is unknown. Something went wrong
	BlockchainPaymentStatusUnknown
)

// CreatePaymentRequestReq is the request for PaymentService.CreatePaymentRequest
type CreatePaymentRequestReq struct {
	IssuerDID   w3c.DID
	UserDID     w3c.DID
	OptionID    uuid.UUID
	SchemaID    uuid.UUID
	Description string
}

// PaymentService is the interface implemented by the payment service
type PaymentService interface {
	CreatePaymentRequest(ctx context.Context, req *CreatePaymentRequestReq) (*domain.PaymentRequest, error)
	GetPaymentRequests(ctx context.Context, issuerDID *w3c.DID) ([]domain.PaymentRequest, error)
	GetPaymentRequest(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentRequest, error)
	DeletePaymentRequest(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) error
	CreatePaymentRequestForProposalRequest(ctx context.Context, proposalRequest *protocol.CredentialsProposalRequestMessage) (*comm.BasicMessage, error)
	GetSettings() payments.Config
	VerifyPayment(ctx context.Context, issuerDID w3c.DID, nonce *big.Int, txHash *string, userDID *w3c.DID) (BlockchainPaymentStatus, error)
	CreatePaymentOption(ctx context.Context, issuerDID *w3c.DID, name, description string, config *domain.PaymentOptionConfig) (uuid.UUID, error)
	GetPaymentOptions(ctx context.Context, issuerDID *w3c.DID) ([]domain.PaymentOption, error)
	GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error)
	DeletePaymentOption(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) error
	UpdatePaymentOption(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID, name, description *string, config *domain.PaymentOptionConfig) error
}
