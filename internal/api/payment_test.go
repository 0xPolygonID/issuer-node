package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

const paymentOptionConfigurationTesting = `
{
  "Config": [
    {
      "paymentOptionId": "1",
      "amount": "500000000000000000",
      "Recipient": "0x1..",
      "SigningKeyId": "pubId"
    },
    {
      "paymentOptionId": "2",
      "amount": "1500000000000000000",
      "Recipient": "0x2..",
      "SigningKeyId": "pubId"
    }
  ]
}
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

	var config PaymentOptionConfig
	var domainConfig domain.PaymentOptionConfig

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
	require.NoError(t, json.Unmarshal([]byte(paymentOptionConfigurationTesting), &domainConfig))

	optionID, err := server.Services.payments.CreatePaymentOption(ctx, issuerDID, "1 POL Payment", "Payment Option explanation", &domainConfig)
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

	require.NoError(t, json.Unmarshal([]byte(paymentOptionConfigurationTesting), &config))

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

	_ = receiverDID

	// Creating an ethereum key
	signingKeyID, err := keyStore.CreateKey(kms.KeyTypeEthereum, issuerDID)
	require.NoError(t, err)

	config := domain.PaymentOptionConfig{
		Config: []domain.PaymentOptionConfigItem{
			{
				PaymentOptionID: 1,
				Amount:          "333",
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
		count    int
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
				UserDID: receiverDID.String(),
				Option:  uuid.New(),
				Credentials: []struct {
					Context string `json:"context"`
					Type    string `json:"type"`
				}{
					{
						Context: "context",
						Type:    "type",
					},
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				msg:      "can't create payment-request: payment option not found",
			},
		},

		{
			name:      "Happy Path",
			auth:      authOk,
			issuerDID: *issuerDID,
			body: CreatePaymentRequestJSONRequestBody{
				UserDID: receiverDID.String(),
				Option:  paymentOptionID,
				Credentials: []struct {
					Context string `json:"context"`
					Type    string `json:"type"`
				}{
					{
						Context: "context",
						Type:    "type",
					},
				},
			},
			expected: expected{
				httpCode: http.StatusCreated,
				count:    10,
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
				raw, err := json.Marshal(response)
				require.NoError(t, err)

				var requestMessage protocol.PaymentRequestMessage
				require.NoError(t, json.Unmarshal(raw, &requestMessage))
				assert.Equal(t, issuerDID.String(), requestMessage.From)
				assert.Equal(t, receiverDID.String(), requestMessage.To)
				assert.Len(t, requestMessage.Body.Payments, 4)

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

// TODO: Review this test!!
func TestServer_VerifyPayment(t *testing.T) {
	/*
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

		receiverDID, err := w3c.ParseDID("did:polygonid:polygon:amoy:2qRYvPBNBTkPaHk1mKBkcLTequfAdsHzXv549ktnL5")
		require.NoError(t, err)

		_ = receiverDID

		// Creating an ethereum key
		signingKeyID, err := keyStore.CreateKey(kms.KeyTypeEthereum, issuerDID)
		require.NoError(t, err)

		// Creating a payment config using previously created key
		config := domain.PaymentOptionConfig{
			Chains: []domain.PaymentOptionConfigChain{
				{
					ChainId:      1101,
					Recipient:    "0x1101...",
					SigningKeyID: signingKeyID.ID,
					Iden3PaymentRailsRequestV1: &domain.PaymentOptionConfigChainIden3PaymentRailsRequestV1{
						Amount:   0.5,
						Currency: "ETH",
					},
					Iden3PaymentRailsERC20RequestV1: nil,
				},
				{
					ChainId:      137,
					Recipient:    "0x137...",
					SigningKeyID: signingKeyID.ID,
					Iden3PaymentRailsRequestV1: &domain.PaymentOptionConfigChainIden3PaymentRailsRequestV1{
						Amount:   0.01,
						Currency: "POL",
					},
					Iden3PaymentRailsERC20RequestV1: &domain.PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1{
						USDT: struct {
							Amount float64 `json:"Amount"`
						}{
							Amount: 5.2,
						},
						USDC: struct {
							Amount float64 `json:"Amount"`
						}{
							Amount: 4.3,
						},
					},
				},
			},
		}

		paymentOptionID, err := server.Services.payments.CreatePaymentOption(ctx, issuerDID, "Cinema ticket single", "Payment Option explanation", &config)
		require.NoError(t, err)

		type expected struct {
			httpCode int
			msg      string
		}

		for _, tc := range []struct {
			name            string
			issuerDID       w3c.DID
			auth            func() (string, string)
			PaymentOptionID uuid.UUID
			body            protocol.PaymentMessage
			expected        expected
		}{
			{
				name:            "Happy Path",
				auth:            authOk,
				issuerDID:       *issuerDID,
				PaymentOptionID: paymentOptionID,
				body: protocol.PaymentMessage{
					ID:       uuid.New().String(),
					Typ:      "application/iden3comm-plain-json",
					Type:     "https://iden3-communication.io/credentials/0.1/payment",
					ThreadID: uuid.New().String(),
					From:     "did:iden3:polygon:mumbai:x3HstHLj2rTp6HHXk2WczYP7w3rpCsRbwCMeaQ2H2",
					To:       "did:polygonid:polygon:mumbai:2qJUZDSCFtpR8QvHyBC4eFm6ab9sJo5rqPbcaeyGC4",
					Body: protocol.PaymentMessageBody{
						Payments: []protocol.Payment{protocol.NewPaymentRails(protocol.Iden3PaymentRailsV1{
							Nonce:   "123",
							Type:    "Iden3PaymentRailsV1",
							Context: protocol.NewPaymentContextString("https://schema.iden3.io/core/jsonld/payment.jsonld"),
							PaymentData: struct {
								TxID    string `json:"txId"`
								ChainID string `json:"chainId"`
							}{
								TxID:    "0x123",
								ChainID: "137",
							},
						})},
					},
				},
				expected: expected{
					httpCode: http.StatusOK,
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				rr := httptest.NewRecorder()
				payload, err := json.Marshal(tc.body)
				require.NoError(t, err)
				url := fmt.Sprintf("/v2/payment/verify/%s", tc.PaymentOptionID)
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
				assert.NoError(t, err)
				req.SetBasicAuth(tc.auth())

				handler.ServeHTTP(rr, req)
				require.Equal(t, tc.expected.httpCode, rr.Code)

				switch tc.expected.httpCode {
				case http.StatusCreated:
					var response VerifyPayment200JSONResponse
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

				case http.StatusBadRequest:
					var response VerifyPayment400JSONResponse
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
					assert.Equal(t, tc.expected.msg, response.Message)
				case http.StatusInternalServerError:
					var response VerifyPayment500JSONResponse
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
					assert.Equal(t, tc.expected.msg, response.Message)
				}
			})
		}

	*/
}
