package repositories

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/driver-did-iden3/pkg/document"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

// CreatePaymentRequest creates a new payment request fixture to be used on tests
func (f *Fixture) CreatePaymentRequest(t *testing.T, issuerDID, recipientDID w3c.DID, paymentOptionID uuid.UUID, nPayments int) *domain.PaymentRequest {
	t.Helper()

	var paymentList []domain.PaymentRequestItem
	paymentRequestID := uuid.New()
	for i := 0; i < nPayments; i++ {
		nonce := big.NewInt(rand.Int63())
		paymentList = append(paymentList, domain.PaymentRequestItem{
			ID:               uuid.New(),
			Nonce:            *nonce,
			PaymentRequestID: paymentRequestID,
			PaymentOptionID:  payments.OptionConfigIDType(i + 1),
			Payment: protocol.Iden3PaymentRailsRequestV1{
				Nonce: "25",
				Type:  protocol.Iden3PaymentRailsRequestV1Type,
				Context: protocol.NewPaymentContextString(
					"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsRequestV1",
					"https://w3id.org/security/suites/eip712sig-2021/v1",
				),
				Recipient:      "0xaddress",
				Amount:         "100",
				ExpirationDate: "ISO string",
				Proof: protocol.PaymentProof{
					protocol.EthereumEip712Signature2021{
						Type:               document.EthereumEip712SignatureProof2021Type,
						ProofPurpose:       "assertionMethod",
						ProofValue:         "0xa05292e9874240c5c2bbdf5a8fefff870c9fc801bde823189fc013d8ce39c7e5431bf0585f01c7e191ea7bbb7110a22e018d7f3ea0ed81a5f6a3b7b828f70f2d1c",
						VerificationMethod: "did:pkh:eip155:0:0x3e1cFE1b83E7C1CdB0c9558236c1f6C7B203C34e#blockchainAccountId",
						Created:            "2024-09-26T12:28:19.702580067Z",
						Eip712: protocol.Eip712Data{
							Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsRequestV1.json",
							PrimaryType: "Iden3PaymentRailsRequestV1",
							Domain: protocol.Eip712Domain{
								Name:              "MCPayment",
								Version:           "1.0.0",
								ChainID:           "0x0",
								VerifyingContract: "0x0000000000000000000000000000000000000000",
							},
						},
					},
				},
				Metadata: "0x",
			},
		})
	}
	paymentRequest := &domain.PaymentRequest{
		ID:              paymentRequestID,
		IssuerDID:       issuerDID,
		RecipientDID:    recipientDID,
		PaymentOptionID: paymentOptionID,
		Payments:        paymentList,
		CreatedAt:       time.Now(),
	}

	_, err := f.paymentRepository.SavePaymentRequest(context.Background(), paymentRequest)
	require.NoError(t, err)
	return paymentRequest
}
