package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

// PaymentOption represents a payment option
type PaymentOption struct {
	ID          uuid.UUID
	IssuerDID   w3c.DID
	Name        string
	Description string
	Config      any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPaymentOption creates a new PaymentOption
func NewPaymentOption(issuerDID w3c.DID, name string, description string, config any) *PaymentOption {
	return &PaymentOption{
		ID:          uuid.New(),
		IssuerDID:   issuerDID,
		Name:        name,
		Description: description,
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
