package repositories

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreatePaymentRequest creates a new payment request fixture to be used on tests
func (f *Fixture) CreatePaymentRequest(t *testing.T, issuerDID, recipientDID w3c.DID, paymentOptionID uuid.UUID, nPayments int) *domain.PaymentRequest {
	t.Helper()

	var payments []domain.PaymentRequestItem
	paymentRequestID := uuid.New()
	for i := 0; i < nPayments; i++ {
		nonce := big.NewInt(rand.Int63())
		payments = append(payments, domain.PaymentRequestItem{
			ID:               uuid.New(),
			Nonce:            *nonce,
			PaymentRequestID: paymentRequestID,
			Payment: protocol.NewPaymentRails(protocol.Iden3PaymentRailsV1{
				Nonce:   nonce.String(),
				Type:    "Iden3PaymentRailsV1",
				Context: protocol.NewPaymentContextString("https://schema.iden3.io/core/jsonld/payment.jsonld"),
				PaymentData: struct {
					TxID    string `json:"txId"`
					ChainID string `json:"chainId"`
				}{
					TxID:    "0x123",
					ChainID: "137",
				},
			}),
		})
	}
	paymentRequest := &domain.PaymentRequest{
		ID:              paymentRequestID,
		IssuerDID:       issuerDID,
		RecipientDID:    recipientDID,
		ThreadID:        "thread-id",
		PaymentOptionID: paymentOptionID,
		Payments:        payments,
		CreatedAt:       time.Now(),
	}

	_, err := f.paymentRepository.SavePaymentRequest(context.Background(), paymentRequest)
	require.NoError(t, err)
	return paymentRequest
}
