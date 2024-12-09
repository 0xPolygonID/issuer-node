package api

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

func TestServer_CreateKey(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateKeyResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		body     CreateKeyRequest
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "no auth header",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should create a key",
			auth: authOk,
			body: CreateKeyRequest{
				KeyType: BJJ,
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "should get an error",
			auth: authOk,
			body: CreateKeyRequest{
				KeyType: "wrong type",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: CreateKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "invalid key type. BJJ and ETH Keys are supported",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys", did)
			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateKey201JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.Id)
			case http.StatusBadRequest:
				var response CreateKey400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_GetKey(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	keyID, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetKeyResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		KeyID    string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "no auth header",
			auth:  authWrong,
			KeyID: keyID.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should get a key",
			auth:  authOk,
			KeyID: keyID.ID,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
		{
			name:  "should get an error",
			auth:  authOk,
			KeyID: "123",
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "the key id can not be decoded from base64",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, tc.KeyID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response GetKey200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.Id)
				assert.Equal(t, keyID, response.Id)
				assert.NotNil(t, response.PublicKey)
				assert.Equal(t, BJJ, response.KeyType)
				assert.False(t, response.IsAuthCredential)
			case http.StatusBadRequest:
				var response GetKey400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_GetKeys(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	_, err = server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	_, err = server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	idenETH, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: ETH})
	require.NoError(t, err)
	didETH, err := w3c.ParseDID(idenETH.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	t.Run("should get an error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys", did)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authWrong())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should get the keys for bjj identity", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys", did)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, uint(3), response.Meta.Total)
		assert.Equal(t, uint(50), response.Meta.MaxResults)
		assert.Equal(t, uint(1), response.Meta.Page)
	})

	t.Run("should get the keys for eth identity", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys", didETH)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, uint(2), response.Meta.Total)
		assert.Equal(t, uint(50), response.Meta.MaxResults)
		assert.Equal(t, uint(1), response.Meta.Page)
	})
}

func TestServer_DeleteKey(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	idenETH, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: ETH})
	require.NoError(t, err)
	didETH, err := w3c.ParseDID(idenETH.Identifier)
	require.NoError(t, err)

	idenETHKeys, _, err := server.keyService.GetAll(ctx, didETH, ports.KeyFilter{MaxResults: 10, Page: 1})
	require.NoError(t, err)
	assert.Equal(t, len(idenETHKeys), 2)

	idenETHBJJKey := ""
	idenETHETHKey := ""
	if idenETHKeys[0].KeyType == kms.KeyTypeBabyJubJub {
		idenETHBJJKey = idenETHKeys[0].KeyID
		idenETHETHKey = idenETHKeys[1].KeyID
	} else {
		idenETHBJJKey = idenETHKeys[1].KeyID
		idenETHETHKey = idenETHKeys[0].KeyID
	}

	keyETHIDToDelete, err := server.keyService.CreateKey(ctx, didETH, kms.KeyTypeEthereum)
	require.NoError(t, err)

	keyID, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	keyIDForAuthCoreClaimID, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	keyIDForAuthCoreClaimIDASByteArr, err := b64.StdEncoding.DecodeString(keyIDForAuthCoreClaimID.ID)
	require.NoError(t, err)

	_, err = server.Services.identity.CreateAuthCredential(ctx, did, string(keyIDForAuthCoreClaimIDASByteArr))
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteKeyResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		did      string
		auth     func() (string, string)
		KeyID    string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "no auth header",
			auth:  authWrong,
			did:   did.String(),
			KeyID: keyID.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should delete the bjj key",
			auth:  authOk,
			did:   did.String(),
			KeyID: keyID.ID,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
		{
			name:  "should delete the eth key",
			auth:  authOk,
			did:   didETH.String(),
			KeyID: keyETHIDToDelete.ID,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
		{
			name:  "should get an error - wrong keyID",
			auth:  authOk,
			did:   did.String(),
			KeyID: "123",
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "the key id can not be decoded from base64",
					},
				},
			},
		},
		{
			name:  "should get an error - associated auth credential is not revoked",
			auth:  authOk,
			did:   did.String(),
			KeyID: keyIDForAuthCoreClaimID.ID,
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "associated auth credential is not revoked",
					},
				},
			},
		},
		{
			name:  "should get an error key is associated with an identity",
			auth:  authOk,
			did:   didETH.String(),
			KeyID: b64.StdEncoding.EncodeToString([]byte(idenETHETHKey)),
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "key is associated with an identity",
					},
				},
			},
		},
		{
			name:  "should get an error associated auth credential is not revoked ",
			auth:  authOk,
			did:   didETH.String(),
			KeyID: b64.StdEncoding.EncodeToString([]byte(idenETHBJJKey)),
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "associated auth credential is not revoked",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys/%s", tc.did, tc.KeyID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response DeleteKey200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			case http.StatusBadRequest:
				var response DeleteKey400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}
