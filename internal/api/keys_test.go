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
			name: "No auth header",
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
						Message: "Invalid key type. BJJ Keys are supported",
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
				assert.NotNil(t, response.KeyID)
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

	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(keyID.ID))

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
			name:  "No auth header",
			auth:  authWrong,
			KeyID: encodedKeyID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should get a key",
			auth:  authOk,
			KeyID: encodedKeyID,
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
						Message: "invalid key id",
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
				assert.NotNil(t, response.KeyID)
				assert.Equal(t, keyID, response.KeyID)
				assert.NotNil(t, response.PublicKey)
				assert.Equal(t, BJJ, response.KeyType)
				assert.False(t, response.IsAuthCoreClaim)
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
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	keyID1, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	keyID2, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	encodedKeyID1 := b64.StdEncoding.EncodeToString([]byte(keyID1.ID))
	encodedKeyID2 := b64.StdEncoding.EncodeToString([]byte(keyID2.ID))

	handler := getHandler(ctx, server)

	type expected struct {
		response GetKeysResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should get the keys",
			auth: authOk,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys", did)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response GetKeys200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response[0].KeyID)
				assert.Equal(t, encodedKeyID1, response[0].KeyID)
				assert.NotNil(t, response[0].PublicKey)
				assert.Equal(t, BJJ, response[0].KeyType)
				assert.False(t, response[0].IsAuthCoreClaim)
				assert.NotNil(t, response[1].KeyID)
				assert.Equal(t, encodedKeyID2, response[1].KeyID)
				assert.NotNil(t, response[1].PublicKey)
				assert.Equal(t, BJJ, response[1].KeyType)
				assert.False(t, response[1].IsAuthCoreClaim)
			case http.StatusBadRequest:
				var response GetKeys400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_DeleteKey(t *testing.T) {
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

	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(keyID.ID))

	keyIDForAuthCoreClaimID, err := server.keyService.CreateKey(ctx, did, kms.KeyTypeBabyJubJub)
	require.NoError(t, err)

	encodedKeyIDForAuthCoreClaimID := b64.StdEncoding.EncodeToString([]byte(keyIDForAuthCoreClaimID.ID))

	_, err = server.Services.identity.AddKey(ctx, did, keyIDForAuthCoreClaimID.ID)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteKeyResponseObject
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
			name:  "No auth header",
			auth:  authWrong,
			KeyID: encodedKeyID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should delete a key",
			auth:  authOk,
			KeyID: encodedKeyID,
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
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "invalid key id",
					},
				},
			},
		},
		{
			name:  "should get an error - key is an auth core claim",
			auth:  authOk,
			KeyID: encodedKeyIDForAuthCoreClaimID,
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: DeleteKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "associated auth core claim is not revoked",
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, tc.KeyID)
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
