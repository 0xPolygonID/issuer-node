package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

const paymentOptionTest = `
{
    "Chains": [
      {
        "ChainId": 137,
        "Recipient": "0x..",
        "SigningKeyId": "<key id>",
        "Iden3PaymentRailsRequestV1": {
          "Amount": "0.01",
          "Currency": "POL" 
        },
        "Iden3PaymentRailsERC20RequestV1": {
          "USDT": {
            "Amount": "3"
          },
          "USDC": {
            "Amount": "3"
          }
        }
      },
      {
        "ChainId": 1101,
        "Recipient": "0x..",
        "SigningKeyId": "<key id>",
        "Iden3PaymentRailsRequestV1": {
          "Amount": "0.5",
          "Currency": "ETH"
        }
      }
    ]
  }
`

func TestPayment_SavePaymentOption(t *testing.T) {
	var paymentOptionConfig domain.PaymentOptionConfig
	ctx := context.Background()

	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:iden3:privado:main:2Sh93vMXNar5fP5ifutHerL9bdUkocB464n3TG6BWV")
	require.NoError(t, err)
	issuerDIDOther, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qYQdd1yqFyrM9ZPqYTE4WHAQH2PX5Rjtj7YDYPppj")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig))

	repo := NewPayment(*storage)
	t.Run("Save payment option", func(t *testing.T) {
		id, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name", "description", &paymentOptionConfig))
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
	})
	t.Run("Save payment option linked to non existing issuer", func(t *testing.T) {
		id, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerDIDOther, "name 2", "description 2", &paymentOptionConfig))
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})
}

func TestPayment_GetAllPaymentOptions(t *testing.T) {
	var paymentOptionConfig domain.PaymentOptionConfig
	ctx := context.Background()
	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:iden3:privado:main:2SbDGSG2TTN1N1UuFaFq7EoFK3RY5wfcotuD8rDCn2")
	require.NoError(t, err)
	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	issuerDIDOther, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qYQdd1yqFyrM9ZPqYTE4WHAQH2PX5Rjtj7YDYPppj")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig))

	repo := NewPayment(*storage)
	ids := make([]uuid.UUID, 0)
	now := time.Now()
	for i := 0; i < 10; i++ {
		id, err := repo.SavePaymentOption(ctx, &domain.PaymentOption{
			ID:          uuid.New(),
			IssuerDID:   *issuerID,
			Name:        fmt.Sprintf("name %d", i),
			Description: fmt.Sprintf("description %d", i),
			Config:      paymentOptionConfig,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		now = now.Add(1 * time.Second)

		require.NoError(t, err)
		ids = append([]uuid.UUID{id}, ids...)
	}
	t.Run("Get all payment options", func(t *testing.T) {
		opts, err := repo.GetAllPaymentOptions(ctx, *issuerID)
		assert.NoError(t, err)
		assert.Len(t, opts, 10)
		for i, opt := range opts {
			assert.Equal(t, ids[i], opt.ID)
			assert.Equal(t, fmt.Sprintf("name %d", len(opts)-i-1), opt.Name)
			assert.Equal(t, fmt.Sprintf("description %d", len(opts)-i-1), opt.Description)
			assert.Equal(t, paymentOptionConfig, opt.Config)
		}
	})
	t.Run("Get all payment options linked to non existing issuer", func(t *testing.T) {
		opts, err := repo.GetAllPaymentOptions(ctx, *issuerDIDOther)
		require.NoError(t, err)
		assert.Len(t, opts, 0)
	})
}

func TestPayment_GetPaymentOptionByID(t *testing.T) {
	var paymentOptionConfig domain.PaymentOptionConfig
	ctx := context.Background()
	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qWxoum8UEJzbUL1Ej9UWjGYHL8oL31BBLJ4ob8bmM")
	require.NoError(t, err)
	issuerDIDOther, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qYQdd1yqFyrM9ZPqYTE4WHAQH2PX5Rjtj7YDYPppj")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig))

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	repo := NewPayment(*storage)
	id, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name", "description", &paymentOptionConfig))
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	t.Run("Get payment option", func(t *testing.T) {
		opt, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
		assert.NoError(t, err)
		assert.Equal(t, id, opt.ID)
		assert.Equal(t, "name", opt.Name)
		assert.Equal(t, "description", opt.Description)
		assert.Equal(t, paymentOptionConfig, opt.Config)
	})
	t.Run("Get payment option linked to non existing issuer", func(t *testing.T) {
		opt, err := repo.GetPaymentOptionByID(ctx, issuerDIDOther, id)
		require.Error(t, err)
		assert.Nil(t, opt)
	})
}

func TestPayment_DeletePaymentOption(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:iden3:privado:main:2Se8ZgrJDWycoKfH9JkBsCuEF127n3nk4G4YW7Dxjo")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	repo := NewPayment(*storage)
	id, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name", "description", &domain.PaymentOptionConfig{}))
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	opt, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
	assert.NoError(t, err)
	assert.Equal(t, id, opt.ID)

	require.NoError(t, repo.DeletePaymentOption(ctx, *issuerID, id))

	opt, err = repo.GetPaymentOptionByID(ctx, issuerID, id)
	assert.Error(t, err)
	assert.Nil(t, opt)
}

func TestPayment_SavePaymentRequest(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qQHtVzfwjywqosK24XT7um3R1Ym5L1GJTbijjcxMq")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	repo := NewPayment(*storage)

	payymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name", "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)

	t.Run("Save payment to not existing payment option id", func(t *testing.T) {
		id, err := repo.SavePaymentRequest(ctx, &domain.PaymentRequest{
			ID:              uuid.New(),
			IssuerDID:       *issuerID,
			RecipientDID:    *issuerID,
			Description:     "this is a payment for cinema",
			Credentials:     []protocol.PaymentRequestInfoCredentials{{Context: "context", Type: "type"}},
			PaymentOptionID: uuid.New(),
			Payments:        []domain.PaymentRequestItem{},
			CreatedAt:       time.Now(),
		})
		assert.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
	})
	t.Run("Save payment with empty Payments", func(t *testing.T) {
		id, err := repo.SavePaymentRequest(ctx, &domain.PaymentRequest{
			ID:              uuid.New(),
			IssuerDID:       *issuerID,
			RecipientDID:    *issuerID,
			Description:     "this is a payment for cinema",
			Credentials:     []protocol.PaymentRequestInfoCredentials{{Context: "context", Type: "type"}},
			PaymentOptionID: payymentOptionID,
			Payments:        nil,
			CreatedAt:       time.Now(),
		})
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
	})
	t.Run("Save payment Happy path", func(t *testing.T) {
		var payments []domain.PaymentRequestItem
		paymentRequestID := uuid.New()
		for i := 0; i < 10; i++ {
			nonce := big.NewInt(rand.Int63())
			payments = append(payments, domain.PaymentRequestItem{
				ID:               uuid.New(),
				Nonce:            *nonce,
				PaymentRequestID: paymentRequestID,
				Payment: protocol.Iden3PaymentRailsERC20RequestV1{
					Nonce:     "123",
					Type:      protocol.Iden3PaymentRailsERC20RequestV1Type,
					Context:   protocol.NewPaymentContextString("https://schema.iden3.io/core/jsonld/payment.jsonld"),
					Recipient: "did1",
					Amount:    "11",
					// etc...
				},
			},
			)
		}
		id, err := repo.SavePaymentRequest(ctx, &domain.PaymentRequest{
			ID:              paymentRequestID,
			IssuerDID:       *issuerID,
			RecipientDID:    *issuerID,
			Description:     "this is a payment for cinema",
			Credentials:     []protocol.PaymentRequestInfoCredentials{{Context: "context", Type: "type"}},
			PaymentOptionID: payymentOptionID,
			Payments:        payments,
			CreatedAt:       time.Now(),
		})
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
	})
}

// TestPayment_GetPaymentRequestByID tests the GetPaymentRequestByID method
func TestPayment_GetPaymentRequestByID(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	repo := NewPayment(*storage)
	issuerID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qSrFwSvp8rLicBhUz4D21nZavGsZufBpjazwQHKmS")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	paymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name", "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)
	expected := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, paymentOptionID, 10)

	t.Run("Get payment request", func(t *testing.T) {
		paymentRequest, err := repo.GetPaymentRequestByID(ctx, expected.IssuerDID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, paymentRequest.ID)
		assert.Equal(t, expected.IssuerDID, paymentRequest.IssuerDID)
		assert.Equal(t, expected.RecipientDID, paymentRequest.RecipientDID)
		assert.Equal(t, expected.PaymentOptionID, paymentRequest.PaymentOptionID)
		assert.InDelta(t, expected.CreatedAt.UnixMilli(), paymentRequest.CreatedAt.UnixMilli(), 1)
		assert.Len(t, expected.Payments, len(paymentRequest.Payments))
		for i, expectedPayment := range expected.Payments {
			assert.Equal(t, expectedPayment.ID, paymentRequest.Payments[i].ID)
			assert.Equal(t, expectedPayment.Nonce, paymentRequest.Payments[i].Nonce)
			assert.Equal(t, expectedPayment.PaymentRequestID, paymentRequest.Payments[i].PaymentRequestID)
			assert.Equal(t, expectedPayment.Payment, paymentRequest.Payments[i].Payment)
		}
	})
}
