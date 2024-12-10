package domain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/payments"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	ID              uuid.UUID
	Credentials     []protocol.PaymentRequestInfoCredentials
	Description     string
	IssuerDID       w3c.DID
	RecipientDID    w3c.DID
	PaymentOptionID uuid.UUID
	Payments        []PaymentRequestItem
	CreatedAt       time.Time
}

// PaymentRequestItem represents a payment request item
type PaymentRequestItem struct {
	ID               uuid.UUID
	Nonce            big.Int
	PaymentRequestID uuid.UUID
	PaymentOptionID  payments.OptionConfigIDType // The numeric id that identify a payment option in payments config file.
	Payment          protocol.PaymentRequestInfoDataItem
}

// PaymentOption represents a payment option
type PaymentOption struct {
	ID          uuid.UUID
	IssuerDID   w3c.DID
	Name        string
	Description string
	Config      PaymentOptionConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPaymentOption creates a new PaymentOption
func NewPaymentOption(issuerDID w3c.DID, name string, description string, config *PaymentOptionConfig) *PaymentOption {
	return &PaymentOption{
		ID:          uuid.New(),
		IssuerDID:   issuerDID,
		Name:        name,
		Description: description,
		Config:      *config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// PaymentOptionConfig represents the configuration of a payment option
type PaymentOptionConfig struct {
	Config []PaymentOptionConfigItem `json:"Config"`
}

// PaymentOptionConfigItem is an item in Payment option config
type PaymentOptionConfigItem struct {
	PaymentOptionID payments.OptionConfigIDType `json:"paymentOptionId"`
	Amount          big.Int                     `json:"amount"`
	Recipient       common.Address              `json:"Recipient"`
	SigningKeyID    string                      `json:"SigningKeyID"`
}
