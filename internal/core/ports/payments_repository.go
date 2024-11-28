package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// PaymentRepository is the interface that defines the available methods for the Payment repository
type PaymentRepository interface {
	SavePaymentOption(ctx context.Context, opt *domain.PaymentOption) (uuid.UUID, error)
	GetAllPaymentOptions(ctx context.Context, issuerDID w3c.DID) ([]domain.PaymentOption, error)
	GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error)
	DeletePaymentOption(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) error
}
