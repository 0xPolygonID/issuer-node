package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

const paymentOptionConfigurationTesting = `
 [
    {
      "paymentOptionId": 1,
      "amount": "500000000000000000",
      "Recipient": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
      "SigningKeyId": "pubId"
    },
    {
      "paymentOptionId": 2,
      "amount": "1500000000000000000",
      "Recipient": "0x53d284357ec70cE289D6D64134DfAc8E511c8a3D",
      "SigningKeyId": "pubId"
    }
]
`

func TestServer_GetPaymentSettings(t *testing.T) {
	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/v2/payment/settings", nil)
	assert.NoError(t, err)
	req.SetBasicAuth(authOk())

	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var response GetPaymentSettings200JSONResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
}

func TestServer_CreatePaymentOption(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)

	var config PaymentOptionConfig
	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	otherDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal([]byte(paymentOptionConfigurationTesting), &config))

	type expected struct {
		httpCode int
		msg      string
	}

	for _, tc := range []struct {
		name      string
		issuerDID w3c.DID
		auth      func() (string, string)
		body      CreatePaymentOptionJSONRequestBody
		expected  expected
	}{
		{
			name:      "no auth header",
			auth:      authWrong,
			issuerDID: *issuerDID,
			body: CreatePaymentOptionJSONRequestBody{
				Config:      config,
				Description: "Payment Option explanation",
				Name:        "1 POL Payment",
			},
			expected: expected{
				httpCode: http.StatusUnauthorized,
				msg:      "Unauthorized",
			},
		},
		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			body: CreatePaymentOptionJSONRequestBody{
				Config:      config,
				Description: "Payment Option explanation",
				Name:        "1 POL Payment",
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
		{
			name:      "Not existing issuerDID",
			auth:      authOk,
			issuerDID: *otherDID,
			body: CreatePaymentOptionJSONRequestBody{
				Config:      config,
				Description: "Payment Option explanation",
				Name:        "1 POL Payment",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "invalid issuer did",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			payload, err := json.Marshal(tc.body)
			require.NoError(t, err)
			url := fmt.Sprintf("/v2/identities/%s/payment/options", tc.issuerDID.String())
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
			assert.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreatePaymentOption201JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			case http.StatusBadRequest:
				var response CreatePaymentOption400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)

			}
		})
	}
}

func TestServer_GetPaymentOption(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)

	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	otherDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
	require.NoError(t, err)

	var config PaymentOptionConfig
	require.NoError(t, json.Unmarshal([]byte(paymentOptionConfigurationTesting), &config))
	domainConfig := domain.PaymentOptionConfig{}
	for _, item := range config {
		amount, ok := new(big.Int).SetString(item.Amount, 10)
		require.True(t, ok)
		domainConfig.Config = append(domainConfig.Config, domain.PaymentOptionConfigItem{
			PaymentOptionID: payments.OptionConfigIDType(item.PaymentOptionID),
			Amount:          *amount,
			Recipient:       common.HexToAddress(item.Recipient),
			SigningKeyID:    item.SigningKeyID,
		})
	}
	optionID, err := server.Services.payments.CreatePaymentOption(
		ctx,
		issuerDID,
		"1 POL Payment",
		"Payment Option explanation",
		&domainConfig)
	require.NoError(t, err)

	type expected struct {
		httpCode int
		msg      string
		option   PaymentOption
	}

	for _, tc := range []struct {
		name      string
		issuerDID w3c.DID
		optionID  uuid.UUID
		auth      func() (string, string)
		expected  expected
	}{
		{
			name:      "no auth header",
			auth:      authWrong,
			issuerDID: *issuerDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusOK,
				option: PaymentOption{
					Id:          optionID,
					IssuerDID:   issuerDID.String(),
					Name:        "1 POL Payment",
					Description: "Payment Option explanation",
					Config:      config,
				},
			},
		},
		{
			name:      "Not existing issuerDID",
			auth:      authOk,
			issuerDID: *otherDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusNotFound,
				msg:      "payment option not found",
			},
		},
		{
			name:      "Not existing Payment option",
			auth:      authOk,
			issuerDID: *issuerDID,
			optionID:  uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				msg:      "payment option not found",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/payment/options/%s", tc.issuerDID.String(), tc.optionID.String())
			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetPaymentOption200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.optionID, response.Id)
				assert.Equal(t, tc.expected.option.Name, response.Name)
				assert.Equal(t, tc.expected.option.Description, response.Description)
				assert.Equal(t, tc.expected.option.IssuerDID, response.IssuerDID)
				assert.Equal(t, tc.expected.option.Config, response.Config)

			case http.StatusNotFound:
				var response GetPaymentOption404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			}
		})
	}
}

func TestServer_GetPaymentOptions(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)

	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	otherDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
	require.NoError(t, err)

	config := domain.PaymentOptionConfig{
		Config: []domain.PaymentOptionConfigItem{
			{
				PaymentOptionID: 1,
				Amount:          *big.NewInt(500000000000000000),
				Recipient:       common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e"),
				SigningKeyID:    "pubId",
			},
			{
				PaymentOptionID: 2,
				Amount:          *big.NewInt(1500000000000000000),
				Recipient:       common.HexToAddress("0x53d284357ec70cE289D6D64134DfAc8E511c8a3D"),
				SigningKeyID:    "pubId",
			},
		},
	}

	for i := 0; i < 10; i++ {
		_, err = server.Services.payments.CreatePaymentOption(ctx, issuerDID, fmt.Sprintf("Payment Option %d", i+1), "Payment Option explanation", &config)
		require.NoError(t, err)
	}
	type expected struct {
		httpCode int
		msg      string
		count    int
	}
	for _, tc := range []struct {
		name      string
		issuerDID w3c.DID
		auth      func() (string, string)
		expected  expected
	}{
		{
			name:      "no auth header",
			auth:      authWrong,
			issuerDID: *issuerDID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			expected: expected{
				httpCode: http.StatusOK,
				count:    10,
			},
		},
		{
			name:      "Other issuer DID with no payment options. Should return empty string",
			auth:      authOk,
			issuerDID: *otherDID,
			expected: expected{
				httpCode: http.StatusOK,
				count:    0,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/payment/options", tc.issuerDID.String())
			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetPaymentOptions200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.count, len(response.Items)) // Check that 10 items are returned
				assert.Equal(t, 1, int(response.Meta.Page))
				assert.Equal(t, tc.expected.count, int(response.Meta.Total))
			case http.StatusBadRequest:
				var response GetPaymentOptions400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			}
		})
	}
}

func TestServer_DeletePaymentOption(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)

	var config domain.PaymentOptionConfig
	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	otherDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
	require.NoError(t, err)

	optionID, err := server.Services.payments.CreatePaymentOption(ctx, issuerDID, "1 POL Payment", "Payment Option explanation", &config)
	require.NoError(t, err)
	type expected struct {
		httpCode int
		msg      string
	}
	for _, tc := range []struct {
		name      string
		issuerDID w3c.DID
		optionID  uuid.UUID
		auth      func() (string, string)
		expected  expected
	}{
		{
			name:      "no auth header",
			auth:      authWrong,
			issuerDID: *issuerDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusOK,
				msg:      "deleted",
			},
		},
		{
			name:      "Not existing issuerDID",
			auth:      authOk,
			issuerDID: *otherDID,
			optionID:  optionID,
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "payment option not found",
			},
		},
		{
			name:      "Not existing Payment option",
			auth:      authOk,
			issuerDID: *issuerDID,
			optionID:  uuid.New(),
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "payment option not found",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/payment/options/%s", tc.issuerDID.String(), tc.optionID.String())
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response DeletePaymentOption200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			case http.StatusBadRequest:
				var response DeletePaymentOption400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			case http.StatusInternalServerError:
				var response DeletePaymentOption500JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			}
		})
	}
}

func TestServer_CreatePaymentRequest(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	receiverDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
	require.NoError(t, err)

	iReq := ports.NewImportSchemaRequest(url, schemaType, nil, "1.0", nil)
	schema, err := server.schemaService.ImportSchema(ctx, *issuerDID, iReq)
	require.NoError(t, err)

	// Creating an ethereum key
	signingKeyID, err := keyStore.CreateKey(kms.KeyTypeEthereum, issuerDID)
	require.NoError(t, err)

	amount := new(big.Int).SetUint64(500000000000000000)
	config := domain.PaymentOptionConfig{
		Config: []domain.PaymentOptionConfigItem{
			{
				PaymentOptionID: 1,
				Amount:          *amount,
				Recipient:       common.Address{},
				SigningKeyID:    signingKeyID.ID,
			},
			{
				PaymentOptionID: 2,
				Amount:          *amount,
				Recipient:       common.Address{},
				SigningKeyID:    signingKeyID.ID,
			},
		},
	}

	paymentOptionID, err := server.Services.payments.CreatePaymentOption(ctx, issuerDID, "Cinema ticket single", "Payment Option explanation", &config)
	require.NoError(t, err)
	type expected struct {
		httpCode int
		msg      string
		resp     CreatePaymentRequestResponse
	}
	for _, tc := range []struct {
		name      string
		issuerDID w3c.DID
		auth      func() (string, string)
		body      CreatePaymentRequestJSONRequestBody
		expected  expected
	}{
		{
			name:      "no auth header",
			auth:      authWrong,
			issuerDID: *issuerDID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Empty body",
			auth:      authOk,
			issuerDID: *issuerDID,
			body:      CreatePaymentRequestJSONRequestBody{},
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "invalid userDID",
			},
		},
		{
			name:      "Not existing payment option",
			auth:      authOk,
			issuerDID: *issuerDID,
			body: CreatePaymentRequestJSONRequestBody{
				UserDID:  receiverDID.String(),
				OptionID: uuid.New(),
				SchemaID: schema.ID,
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "can't create payment-request: failed to get payment option: payment option not found",
			},
		},
		{
			name:      "Not existing schema",
			auth:      authOk,
			issuerDID: *issuerDID,
			body: CreatePaymentRequestJSONRequestBody{
				UserDID:  receiverDID.String(),
				OptionID: paymentOptionID,
				SchemaID: uuid.New(),
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "can't create payment-request: failed to get schema: schema not found",
			},
		},
		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			body: CreatePaymentRequestJSONRequestBody{
				UserDID:     receiverDID.String(),
				OptionID:    paymentOptionID,
				SchemaID:    schema.ID,
				Description: "Payment Request",
			},
			expected: expected{
				httpCode: http.StatusCreated,
				resp: CreatePaymentRequestResponse{
					CreatedAt:       time.Now(),
					IssuerDID:       issuerDID.String(),
					PaymentOptionID: paymentOptionID,
					Payments: []PaymentRequestInfo{
						{
							Credentials: []protocol.PaymentRequestInfoCredentials{
								{
									Type:    "KYCCountryOfResidenceCredential",
									Context: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
								},
							},
							Description: "lala",
							Data:        protocol.PaymentRequestInfoData{},
						},
					},
					RecipientDID: receiverDID.String(),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			payload, err := json.Marshal(tc.body)
			require.NoError(t, err)
			url := fmt.Sprintf("/v2/identities/%s/payment-request", tc.issuerDID.String())
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
			assert.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreatePaymentRequest201JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotEqual(t, uuid.Nil, response.Id)
				assert.Equal(t, tc.expected.resp.IssuerDID, response.IssuerDID)
				assert.Equal(t, tc.expected.resp.RecipientDID, response.RecipientDID)
				assert.InDelta(t, time.Now().UnixMilli(), response.CreatedAt.UnixMilli(), 10)
				/*
					assert.Equal(t, len(tc.expected.resp.Payments), len(response.Payments))
					for i := range tc.expected.resp.Payments {
						assert.NotEqual(t, big.Int{}, response.Payments[i].Nonce)
						assert.NotEqual(t, uuid.Nil, response.Payments[i].PaymentRequestID)
						assert.NotEqual(t, uuid.Nil, response.Payments[i].Id)
						// TODO: Fix it assert.Equal(t, tc.expected.resp.Payments[i].Payment, response.Payments[i].Payment)
					}

				*/
			case http.StatusBadRequest:
				var response CreatePaymentRequest400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			case http.StatusInternalServerError:
				var response CreatePaymentRequest500JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.msg, response.Message)
			}
		})
	}
}
