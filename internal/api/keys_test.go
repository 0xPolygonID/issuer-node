package api

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

func TestServer_CreateKey(t *testing.T) {
	const (
		method     = "iden3"
		blockchain = "privado"
		network    = "main"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "http://issuer-node", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	idenETH, err := server.Services.identity.Create(ctx, "http://issuer-node", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: ETH})
	require.NoError(t, err)
	didETH, err := w3c.ParseDID(idenETH.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateKeyResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		body     CreateKeyRequest
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "no auth header",
			did:  did.String(),
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should create a bjj key",
			auth: authOk,
			did:  did.String(),
			body: CreateKeyRequest{
				KeyType: CreateKeyRequestKeyType(KeyKeyTypeBabyjubJub),
				Name:    "my-bjj-key",
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "should get an error",
			auth: authOk,
			did:  did.String(),
			body: CreateKeyRequest{
				KeyType: "wrong type",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: CreateKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "invalid key type. babyjujJub and secp256k1 keys are supported are supported",
					},
				},
			},
		},
		{
			name: "should get an error - empty name",
			auth: authOk,
			did:  did.String(),
			body: CreateKeyRequest{
				KeyType: CreateKeyRequestKeyType(KeyKeyTypeBabyjubJub),
				Name:    "",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: CreateKey400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "name is required",
					},
				},
			},
		},
		{
			name: "should create an eth key",
			auth: authOk,
			did:  didETH.String(),
			body: CreateKeyRequest{
				KeyType: CreateKeyRequestKeyType(KeyKeyTypeSecp256k1),
				Name:    "my-eth-key",
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/keys", tc.did)
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
		method     = "iden3"
		blockchain = "privado"
		network    = "main"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "my-key")
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
				assert.Equal(t, KeyKeyTypeBabyjubJub, response.KeyType)
				assert.False(t, response.IsAuthCredential)
				assert.True(t, "my-key" == response.Name)
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
		method     = "iden3"
		blockchain = "privado"
		network    = "main"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "http://issuer-node", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	idenETH, err := server.Services.identity.Create(ctx, "http://issuer-node", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: ETH})
	require.NoError(t, err)
	didETH, err := w3c.ParseDID(idenETH.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	keys, err := keyStore.KeysByIdentity(ctx, *did)
	require.NoError(t, err)
	assert.Len(t, keys, 1)
	pubKey, err := keyStore.PublicKey(keys[0])
	require.NoError(t, err)
	publicKey := hexutil.Encode(pubKey)

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
		assert.Equal(t, uint(1), response.Meta.Total)
		assert.Equal(t, uint(50), response.Meta.MaxResults)
		assert.Equal(t, uint(1), response.Meta.Page)
		assert.Equal(t, 1, countAuthCredentials(t, response.Items))
	})

	t.Run("should get the keys for bjj identity with pagination", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			name := fmt.Sprintf("z-key-%s", string('A'+rune(i+1)))
			_, err = server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, name)
			require.NoError(t, err)
		}

		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys?max_results=11&page=1", did)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, uint(21), response.Meta.Total)
		assert.Equal(t, uint(11), response.Meta.MaxResults)
		assert.Equal(t, uint(1), response.Meta.Page)

		rr = httptest.NewRecorder()
		url = fmt.Sprintf("/v2/identities/%s/keys?max_results=11&page=2", did)
		req, err = http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response1 GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response1))
		assert.Equal(t, uint(21), response1.Meta.Total)
		assert.Equal(t, uint(11), response1.Meta.MaxResults)
		assert.Equal(t, uint(2), response1.Meta.Page)

		all := append(response.Items, response1.Items...)
		assert.Equal(t, 1, countAuthCredentials(t, all))

		// Check that the keys are sorted by name
		for i, key := range all {
			assert.Equal(t, string(KeyKeyTypeBabyjubJub), string(key.KeyType))
			if i != 0 {
				assert.Equal(t, fmt.Sprintf("z-key-%s", string('A'+rune(i))), key.Name)
			} else {
				assert.Equal(t, publicKey, key.Name)
			}
		}
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

		assert.Equal(t, 2, countAuthCredentials(t, response.Items))
	})

	t.Run("should get the keys for eth identity with pagination", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			_, err = server.keyService.Create(ctx, didETH, kms.KeyTypeBabyJubJub, fmt.Sprintf("my-key-%d", i))
			require.NoError(t, err)
		}

		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys?max_results=15&page=1", didETH)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, uint(22), response.Meta.Total)
		assert.Equal(t, uint(15), response.Meta.MaxResults)
		assert.Equal(t, uint(1), response.Meta.Page)

		rr = httptest.NewRecorder()
		url = fmt.Sprintf("/v2/identities/%s/keys?max_results=15&page=2", didETH)
		req, err = http.NewRequest(http.MethodGet, url, nil)
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)

		var response1 GetKeys200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response1))
		assert.Equal(t, uint(22), response1.Meta.Total)
		assert.Equal(t, uint(15), response1.Meta.MaxResults)
		assert.Equal(t, uint(2), response1.Meta.Page)

		all := append(response.Items, response1.Items...)
		assert.Equal(t, 2, countAuthCredentials(t, all))
	})
}

func countAuthCredentials(t *testing.T, keys []Key) int {
	t.Helper()
	count := 0
	for _, key := range keys {
		if key.IsAuthCredential {
			count++
		}
	}
	return count
}

func TestServer_DeleteKey(t *testing.T) {
	const (
		method     = "iden3"
		blockchain = "privado"
		network    = "main"
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

	keyETHIDToDelete, err := server.keyService.Create(ctx, didETH, kms.KeyTypeEthereum, "key-eth-to-delete")
	require.NoError(t, err)

	keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "key-bjj-to-delete")
	require.NoError(t, err)

	keyIDForAuthCoreClaimID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "key-bjj-for-auth-core-claim-id")
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

func TestServer_UpdateKey(t *testing.T) {
	const (
		method     = "iden3"
		blockchain = "privado"
		network    = "main"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)

	iden, err := server.Services.identity.Create(ctx, "http://issuer-node", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	t.Run("should get an error - no auth header", func(t *testing.T) {
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, "123")
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, nil))
		require.NoError(t, err)
		req.SetBasicAuth(authWrong())

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should get an error - wrong keyID", func(t *testing.T) {
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, "123123")
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, nil))
		require.NoError(t, err)
		req.SetBasicAuth(authOk())

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var response UpdateKey400JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "the key id can not be decoded from base64", response.Message)
	})

	t.Run("should update a key", func(t *testing.T) {
		keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "my-key")
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, keyID.ID)
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, UpdateKeyJSONRequestBody{
			Name: "new-name",
		}))
		require.NoError(t, err)
		req.SetBasicAuth(authOk())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var response UpdateKey200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "key updated", response.Message)
	})

	t.Run("should get an error - duplicate key name", func(t *testing.T) {
		keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "my-key")
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, keyID.ID)
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, UpdateKeyJSONRequestBody{
			Name: "new-name",
		}))
		require.NoError(t, err)
		req.SetBasicAuth(authOk())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var response UpdateKey400JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "duplicate key name", response.Message)
	})

	t.Run("should get an error - name is required", func(t *testing.T) {
		keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "my-key-to-not-update")
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, keyID.ID)
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, UpdateKeyJSONRequestBody{
			Name: "",
		}))
		require.NoError(t, err)
		req.SetBasicAuth(authOk())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var response UpdateKey400JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "name is required", response.Message)
	})

	t.Run("should get an error - key not found", func(t *testing.T) {
		keyID, err := server.keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "my-key-to-not-update-2")
		require.NoError(t, err)

		decodedKeyID, err := b64.StdEncoding.DecodeString(keyID.ID)
		require.NoError(t, err)

		require.NoError(t, server.keyService.Delete(ctx, did, string(decodedKeyID)))
		rr := httptest.NewRecorder()
		url := fmt.Sprintf("/v2/identities/%s/keys/%s", did, keyID.ID)
		req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, UpdateKeyJSONRequestBody{
			Name: "new-name",
		}))
		require.NoError(t, err)
		req.SetBasicAuth(authOk())
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)
		var response UpdateKey404JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "key not found", response.Message)
	})
}
