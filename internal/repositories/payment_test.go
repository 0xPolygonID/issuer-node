package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

const paymentOptionTest = `
 [
    {
      "paymentOptionId": 1,
      "amount": 500000000000000000,
      "Recipient": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "SigningKeyId": "pubId",
	  "Expiration": "2025-04-17T11:40:43.681857-03:00"
    },
    {
      "paymentOptionId": 2,
      "amount": 1500000000000000000,
      "Recipient": "0x53d284357ec70cE289D6D64134DfAc8E511c8a3D",
      "SigningKeyId": "pubId",
	  "Expiration": "2025-04-17T11:40:43.681857-03:00"
    }
]
`

const paymentOptionTest2 = `
 [
    {
      "paymentOptionId": 3,
      "amount": 400000000000000000,
      "Recipient": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "SigningKeyId": "pubId",
	  "Expiration": "2025-04-17T11:40:43.681857-03:00"
    }
]
`

func TestPayment_SavePaymentOption(t *testing.T) {
	var paymentOptionConfig domain.PaymentOptionConfig
	var paymentOptionConfigToUpdate domain.PaymentOptionConfig
	ctx := context.Background()

	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:iden3:privado:main:2Sh93vMXNar5fP5ifutHerL9bdUkocB464n3TG6BWV")
	require.NoError(t, err)
	issuerDIDOther, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qYQdd1yqFyrM9ZPqYTE4WHAQH2PX5Rjtj7YDYPppj")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig.PaymentOptions))
	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest2), &paymentOptionConfigToUpdate.PaymentOptions))

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

	t.Run("Save existing payment option - new name", func(t *testing.T) {
		paymentOptionName := "payment-option-" + uuid.NewString()
		paymentOption := domain.NewPaymentOption(*issuerID, paymentOptionName, "description", &paymentOptionConfig)
		id, err := repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		paymentOption.ID = id
		paymentOption.Name = "new-name"
		id, err = repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		updatedPaymentOption, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
		assert.NoError(t, err)
		assert.Equal(t, "new-name", updatedPaymentOption.Name)
	})

	t.Run("Save existing payment option - new description", func(t *testing.T) {
		paymentOptionName := "payment-option-" + uuid.NewString()
		paymentOption := domain.NewPaymentOption(*issuerID, paymentOptionName, "description", &paymentOptionConfig)
		id, err := repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		paymentOption.ID = id
		paymentOption.Description = "new-description"
		id, err = repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		updatedPaymentOption, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
		assert.NoError(t, err)
		assert.Equal(t, "new-description", updatedPaymentOption.Description)
	})

	t.Run("Save existing payment option - new config", func(t *testing.T) {
		paymentOptionName := "payment-option-" + uuid.NewString()
		paymentOption := domain.NewPaymentOption(*issuerID, paymentOptionName, "description", &paymentOptionConfig)
		id, err := repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, id)
		paymentOption.ID = id
		paymentOption.Config = paymentOptionConfigToUpdate
		id, err = repo.SavePaymentOption(ctx, paymentOption)
		assert.NoError(t, err)
		updatedPaymentOption, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
		assert.NoError(t, err)
		assert.Equal(t, paymentOptionConfigToUpdate.PaymentOptions[0].PaymentOptionID, updatedPaymentOption.Config.PaymentOptions[0].PaymentOptionID)
		assert.Equal(t, paymentOptionConfigToUpdate.PaymentOptions[0].Amount, updatedPaymentOption.Config.PaymentOptions[0].Amount)
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

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig.PaymentOptions))

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

	require.NoError(t, json.Unmarshal([]byte(paymentOptionTest), &paymentOptionConfig.PaymentOptions))

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	repo := NewPayment(*storage)
	paymentOption := domain.NewPaymentOption(*issuerID, "payment option name "+uuid.NewString(), "description", &paymentOptionConfig)
	id, err := repo.SavePaymentOption(ctx, paymentOption)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, id)

	t.Run("Get payment option", func(t *testing.T) {
		opt, err := repo.GetPaymentOptionByID(ctx, issuerID, id)
		assert.NoError(t, err)
		assert.Equal(t, id, opt.ID)
		assert.Equal(t, paymentOption.Name, opt.Name)
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
	id, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name"+uuid.NewString(), "description", &domain.PaymentOptionConfig{}))
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

	paymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name"+uuid.NewString(), "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)

	t.Run("Save payment to not existing payment option id", func(t *testing.T) {
		id, err := repo.SavePaymentRequest(ctx, &domain.PaymentRequest{
			ID:              uuid.New(),
			IssuerDID:       *issuerID,
			UserDID:         *issuerID,
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
			UserDID:         *issuerID,
			Description:     "this is a payment for cinema",
			Credentials:     []protocol.PaymentRequestInfoCredentials{{Context: "context", Type: "type"}},
			PaymentOptionID: paymentOptionID,
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
				PaymentOptionID: 6,
			},
			)
		}
		id, err := repo.SavePaymentRequest(ctx, &domain.PaymentRequest{
			ID:              paymentRequestID,
			IssuerDID:       *issuerID,
			UserDID:         *issuerID,
			Description:     "this is a payment for cinema",
			Credentials:     []protocol.PaymentRequestInfoCredentials{{Context: "context", Type: "type"}},
			PaymentOptionID: paymentOptionID,
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
	paymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name"+uuid.NewString(), "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)
	expected := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, paymentOptionID, 10, nil)

	t.Run("Get payment request", func(t *testing.T) {
		paymentRequest, err := repo.GetPaymentRequestByID(ctx, expected.IssuerDID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, paymentRequest.ID)
		assert.Equal(t, expected.IssuerDID, paymentRequest.IssuerDID)
		assert.Equal(t, expected.UserDID, paymentRequest.UserDID)
		assert.Equal(t, expected.PaymentOptionID, paymentRequest.PaymentOptionID)
		assert.InDelta(t, expected.CreatedAt.UnixMilli(), paymentRequest.CreatedAt.UnixMilli(), 1)
		assert.Len(t, expected.Payments, len(paymentRequest.Payments))
		for i, expectedPayment := range expected.Payments {
			assert.Equal(t, expectedPayment.ID, paymentRequest.Payments[i].ID)
			assert.Equal(t, expectedPayment.Nonce, paymentRequest.Payments[i].Nonce)
			assert.Equal(t, expectedPayment.PaymentRequestID, paymentRequest.Payments[i].PaymentRequestID)
			assert.Equal(t, expectedPayment.Payment, paymentRequest.Payments[i].Payment)
			assert.NotZero(t, paymentRequest.Payments[i].PaymentOptionID)
		}
	})
}

func TestPayment_GetPaymentRequestItem(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	issuerID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qX7mBeonrp5u7GapztqjYZTdLsv9XBhyXgjYQGjgi")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})
	repo := NewPayment(*storage)

	payymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name"+uuid.NewString(), "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)
	expected := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, payymentOptionID, 10, nil)

	t.Run("Get payment request item", func(t *testing.T) {
		paymentRequestItem, err := repo.GetPaymentRequestItem(ctx, *issuerID, &expected.Payments[0].Nonce)
		require.NoError(t, err)
		assert.Equal(t, expected.Payments[0].ID, paymentRequestItem.ID)
	})
}

func TestPayment_GetManyPaymentRequests(t *testing.T) {
	ctx := context.Background()
	fixture := NewFixture(storage)
	repo := NewPayment(*storage)
	issuerID, err := w3c.ParseDID("did:iden3:privado:main:2SecTLMtYpAGa8dnD3Y6wZQrNvFSyQvcwWZBfNr4Bg")
	require.NoError(t, err)

	fixture.CreateIdentity(t, &domain.Identity{Identifier: issuerID.String()})

	paymentOptionID, err := repo.SavePaymentOption(ctx, domain.NewPaymentOption(*issuerID, "name"+uuid.NewString(), "description", &domain.PaymentOptionConfig{}))
	require.NoError(t, err)
	ids := make([]string, 0)
	for i := 0; i < 10; i++ {
		pr := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, paymentOptionID, 10, nil)
		ids = append(ids, pr.ID.String())
	}

	all, err := repo.GetAllPaymentRequests(ctx, *issuerID, nil)
	require.NoError(t, err)
	require.Len(t, all, len(ids))
	for _, pr := range all {
		assert.True(t, slices.Contains(ids, pr.ID.String()))
	}

	// search by user DID
	userDIDString := "did:iden3:polygon:amoy:xBBTMzDnifT3VucYSUnQt8PyRmVSPKnVEwyBeH6B5"
	userDID, err := w3c.ParseDID(userDIDString)
	require.NoError(t, err)
	prUserSearch := fixture.CreatePaymentRequest(t, *issuerID, *userDID, paymentOptionID, 1, nil)
	userPayment, err := repo.GetAllPaymentRequests(ctx, *issuerID, &domain.PaymentRequestsQueryParams{
		UserDID: &userDIDString,
	})
	require.NoError(t, err)
	require.Len(t, userPayment, 1)
	assert.Equal(t, prUserSearch.ID, userPayment[0].ID)

	// search by schema ID
	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerID,
		URL:       "https://domain.org/this/is/an/url",
		Type:      "schemaType",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
	}
	fixture.CreateSchema(t, ctx, s)
	prSchemaSearch := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, paymentOptionID, 1, &s.ID)
	schemaPayment, err := repo.GetAllPaymentRequests(ctx, *issuerID, &domain.PaymentRequestsQueryParams{
		SchemaID: &s.ID,
	})
	require.NoError(t, err)
	require.Len(t, schemaPayment, 1)
	assert.Equal(t, prSchemaSearch.ID, schemaPayment[0].ID)

	// search by nonce
	prNonceSearch := fixture.CreatePaymentRequest(t, *issuerID, *issuerID, paymentOptionID, 1, nil)
	nonce := prNonceSearch.Payments[0].Nonce.String()
	noncePayment, err := repo.GetAllPaymentRequests(ctx, *issuerID, &domain.PaymentRequestsQueryParams{
		Nonce: &nonce,
	})
	require.NoError(t, err)
	require.Len(t, noncePayment, 1)
	assert.Equal(t, prNonceSearch.ID, noncePayment[0].ID)

	// search by non existed userDID
	noncePayment, err = repo.GetAllPaymentRequests(ctx, *issuerID, &domain.PaymentRequestsQueryParams{
		UserDID: common.ToPointer("did:polygonid:polygon:amoy:2qVXETy9QdEE7KB8BeMCNuYgAjZj8j5CfqURAxvUsN"),
	})
	require.NoError(t, err)
	require.Len(t, noncePayment, 0)

	// search by all params:
	allParamsSearch := fixture.CreatePaymentRequest(t, *issuerID, *userDID, paymentOptionID, 1, &s.ID)
	allParamsPayment, err := repo.GetAllPaymentRequests(ctx, *issuerID, &domain.PaymentRequestsQueryParams{
		UserDID:  &userDIDString,
		SchemaID: &s.ID,
		Nonce:    common.ToPointer(allParamsSearch.Payments[0].Nonce.String()),
	})
	require.NoError(t, err)
	require.Len(t, allParamsPayment, 1)
	assert.Equal(t, allParamsSearch.ID, allParamsPayment[0].ID)
}
