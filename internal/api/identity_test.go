package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_CreateIdentity(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
		ETH        = "ETH"
	)
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  *string
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		input    CreateIdentityRequest
		expected expected
	}

	authBJJCredentialStatus := (CreateIdentityRequestCredentialStatusType)(verifiable.Iden3commRevocationStatusV1)
	authBJJCredentialStatusOnChain := (CreateIdentityRequestCredentialStatusType)(verifiable.Iden3OnchainSparseMerkleTreeProof2023)
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should create a BJJ identity for amoy network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: string(core.Amoy), Type: BJJ},
				CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a ETH identity for amoy network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: string(core.Amoy), Type: ETH},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a BJJ identity",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: BJJ}, DisplayName: common.ToPointer("blockchain display name"), CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should create a ETH identity",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: ETH},
			},
			expected: expected{
				httpCode: 201,
				message:  nil,
			},
		},
		{
			name: "should return an error wrong network",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: "mynetwork", Type: BJJ},
				CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("error getting reverse hash service settings: rhsSettings not found for polygon:mynetwork"),
			},
		},
		{
			name: "should return an error wrong method",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: "my method", Network: network, Type: BJJ},
				CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("cannot create identity: can't add genesis claims to tree: wrong DID Metadata"),
			},
		},
		{
			name: "should return an error wrong blockchain",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: BJJ},
				CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("error getting reverse hash service settings: rhsSettings not found for my blockchain:amoy"),
			},
		},
		{
			name: "should return an error wrong type",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: "a wrong type"},
				CredentialStatusType: &authBJJCredentialStatus,
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("Type must be BJJ or ETH"),
			},
		},
		{
			name: "should return an error wrong auth core claim",
			auth: authOk,
			input: CreateIdentityRequest{
				DidMetadata: struct {
					Blockchain string                               `json:"blockchain"`
					Method     string                               `json:"method"`
					Network    string                               `json:"network"`
					Type       CreateIdentityRequestDidMetadataType `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: BJJ},
				CredentialStatusType: &authBJJCredentialStatusOnChain,
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("Credential Status Type 'Iden3OnchainSparseMerkleTreeProof2023' is not supported by the issuer"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/v2/identities", tests.JSONBody(t, tc.input))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateIdentityResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				require.NotNil(t, response.Identifier)
				assert.Contains(t, *response.Identifier, tc.input.DidMetadata.Network)
				assert.NotNil(t, response.State.CreatedAt)
				assert.NotNil(t, response.State.ModifiedAt)
				assert.NotNil(t, response.State.State)
				assert.NotNil(t, response.State.Status)
				assert.Equal(t, string(verifiable.Iden3commRevocationStatusV1), string(response.CredentialStatusType))
				if tc.input.DidMetadata.Type == BJJ {
					assert.NotNil(t, *response.State.ClaimsTreeRoot)
				}
				if tc.input.DidMetadata.Type == ETH {
					assert.NotNil(t, *response.Address)
				}
			case http.StatusBadRequest:
				var response CreateIdentity400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
	t.Run("Duplicated display name", func(t *testing.T) {
		bodyRequest := CreateIdentityRequest{
			DidMetadata: struct {
				Blockchain string                               `json:"blockchain"`
				Method     string                               `json:"method"`
				Network    string                               `json:"network"`
				Type       CreateIdentityRequestDidMetadataType `json:"type"`
			}{
				Blockchain: blockchain,
				Method:     method,
				Network:    network,
				Type:       BJJ,
			},
			DisplayName: common.ToPointer("Very common display name"),
		}
		// First request
		req, err := http.NewRequest("POST", "/v2/identities", tests.JSONBody(t, bodyRequest))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code) // First time we expect 201

		// Second request
		req, err = http.NewRequest("POST", "/v2/identities", tests.JSONBody(t, bodyRequest))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusConflict, rr.Code) // Second time we expect a conflict 409
		var response CreateIdentity409JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "display name field already exists: <Very common display name>", response.Message)
	})
}

func TestServer_GetIdentities(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	identity1 := &domain.Identity{Identifier: "did:polygonid:polygon:mumbai:2qE1ZT16aqEWhh9mX9aqM2pe2ZwV995dTkReeKwCaQ"}
	identity2 := &domain.Identity{Identifier: "did:polygonid:polygon:mumbai:2qMHFTHn2SC3XkBEJrR4eH4Yk8jRGg5bzYYG1ZGECa"}
	fixture := repositories.NewFixture(storage)
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

	type expected struct {
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
			name: "should return all the entities",
			auth: authOk,
			expected: expected{
				httpCode: 200,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/v2/identities", nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response GetIdentities200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.httpCode, rr.Code)
				assert.True(t, len(response) >= 2)
			}
		})
	}
}

func TestServer_GetIdentityDetails(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	identity, err := server.identityService.Create(ctx, cfg.ServerUrl, &ports.DIDCreationOptions{
		Method:               "polygonid",
		Blockchain:           "polygon",
		Network:              "amoy",
		KeyType:              "BJJ",
		DisplayName:          common.ToPointer("my display name"),
		AuthCredentialStatus: verifiable.Iden3commRevocationStatusV1,
	})
	require.NoError(t, err)

	type expected struct {
		httpCode             int
		displayName          *string
		status               string
		state                string
		identifier           string
		keyType              string
		credentialStatusType verifiable.CredentialStatusType
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			did:  identity.Identifier,
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "identity not found",
			auth: authOk,
			did:  "did:polygonid:polygon:amoy:2qE1ZT16aqEWhh9mX9aqM2pe2ZwV995dTkReeKwCaQ",
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "should retrieve identity details",
			auth: authOk,
			did:  identity.Identifier,
			expected: expected{
				httpCode:             200,
				identifier:           identity.Identifier,
				displayName:          common.ToPointer("my display name"),
				status:               "confirmed",
				state:                "state",
				keyType:              "BJJ",
				credentialStatusType: verifiable.Iden3commRevocationStatusV1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			url := fmt.Sprintf("/v2/identities/%s", tc.did)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response GetIdentityDetails200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.identifier, response.Identifier)
				assert.Equal(t, tc.expected.displayName, response.DisplayName)
				assert.Equal(t, tc.expected.status, response.State.Status)
				assert.NotNil(t, response.State.State)
				assert.NotNil(t, response.State.ClaimsTreeRoot)
				assert.Equal(t, tc.expected.keyType, response.KeyType)
				assert.Equal(t, tc.expected.credentialStatusType, verifiable.CredentialStatusType(response.CredentialStatusType))
				assert.Equal(t, "", *identity.Address)
				assert.Nil(t, identity.Balance)
				assert.Len(t, response.AuthCredentialsIDs, 1)
			}
		})
	}
}

func TestServer_UpdateIdentity(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	identity := &domain.Identity{Identifier: "did:polygonid:polygon:amoy:2qQ8S2VKdQv7xYgzCn7KW2xgzUWrTRQjoZDYavJHBq"}
	fixture := repositories.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	state := domain.IdentityState{
		Identifier: identity.Identifier,
		State:      common.ToPointer("state"),
		Status:     domain.StatusCreated,
		ModifiedAt: time.Now(),
		CreatedAt:  time.Now(),
	}
	fixture.CreateIdentityStatus(t, state)

	type expected struct {
		httpCode    int
		displayName *string
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
				httpCode:    http.StatusUnauthorized,
				displayName: common.ToPointer("new display name"),
			},
		},
		{
			name: "should update display name",
			auth: authOk,
			expected: expected{
				httpCode:    200,
				displayName: common.ToPointer("new display name"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			body := UpdateIdentityJSONBody{
				DisplayName: *tc.expected.displayName,
			}

			url := fmt.Sprintf("/v2/identities/%s", identity.Identifier)
			req, err := http.NewRequest("PATCH", url, tests.JSONBody(t, body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response UpdateIdentity200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			}
		})
	}
}

func TestServer_CreateAuthCredential(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()

	server := newTestServer(t, nil)
	handler := getHandler(ctx, server)

	iden, err := server.Services.identity.Create(ctx, "http://polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	did := iden.Identifier

	issuerDID, err := w3c.ParseDID(did)
	require.NoError(t, err)

	t.Run("no auth header", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, nil))
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should create an auth credential with default values", func(t *testing.T) {
		key, err := server.keyService.Create(ctx, issuerDID, kms.KeyTypeBabyJubJub, uuid.New().String())
		require.NoError(t, err)

		body := CreateAuthCredentialRequest{
			KeyID:                key.ID,
			CredentialStatusType: CreateAuthCredentialRequestCredentialStatusType(verifiable.Iden3commRevocationStatusV1),
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, body))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
		var response CreateAuthCredential201JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.NotNil(t, response.Id)
	})

	t.Run("should get an error - duplicated auth credential", func(t *testing.T) {
		key, err := server.keyService.Create(ctx, issuerDID, kms.KeyTypeBabyJubJub, uuid.New().String())
		require.NoError(t, err)

		body := CreateAuthCredentialRequest{
			KeyID:                key.ID,
			CredentialStatusType: CreateAuthCredentialRequestCredentialStatusType(verifiable.Iden3commRevocationStatusV1),
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, body))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)

		body = CreateAuthCredentialRequest{
			KeyID:                key.ID,
			CredentialStatusType: CreateAuthCredentialRequestCredentialStatusType(verifiable.Iden3commRevocationStatusV1),
		}
		rr = httptest.NewRecorder()
		req, err = http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, body))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var response CreateAuthCredential400JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "hash already exists. This means an auth credential was already created with this key", response.Message)
	})

	t.Run("should get an error - credential status type not supported", func(t *testing.T) {
		key, err := server.keyService.Create(ctx, issuerDID, kms.KeyTypeBabyJubJub, uuid.New().String())
		require.NoError(t, err)

		body := CreateAuthCredentialRequest{
			KeyID:                key.ID,
			CredentialStatusType: CreateAuthCredentialRequestCredentialStatusType(verifiable.Iden3OnchainSparseMerkleTreeProof2023),
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, body))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
		var response CreateAuthCredential400JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, "Credential Status Type 'Iden3OnchainSparseMerkleTreeProof2023' is not supported by the issuer", response.Message)
	})

	t.Run("should create an auth credential", func(t *testing.T) {
		key, err := server.keyService.Create(ctx, issuerDID, kms.KeyTypeBabyJubJub, uuid.New().String())
		require.NoError(t, err)

		authCredentialExpiration := time.Now().UTC().Unix()

		revNonce, err := common.RandInt64()
		require.NoError(t, err)
		body := CreateAuthCredentialRequest{
			KeyID:                key.ID,
			CredentialStatusType: CreateAuthCredentialRequestCredentialStatusType(verifiable.Iden3commRevocationStatusV1),
			RevNonce:             &revNonce,
			Version:              common.ToPointer[uint32](1),
			Expiration:           common.ToPointer[int64](authCredentialExpiration),
		}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", fmt.Sprintf("/v2/identities/%s/create-auth-credential", did), tests.JSONBody(t, body))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
		var response CreateAuthCredential201JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.NotNil(t, response.Id)

		rr = httptest.NewRecorder()
		req, err = http.NewRequest("GET", fmt.Sprintf("/v2/identities/%s/credentials/%s", did, response.Id), tests.JSONBody(t, nil))
		req.SetBasicAuth(authOk())
		require.NoError(t, err)
		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var response2 GetCredential200JSONResponse
		assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response2))
		assert.NotNil(t, response2.Id)
		status, ok := response2.Vc.CredentialStatus.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(revNonce), status["revocationNonce"])
		assert.Equal(t, string(verifiable.Iden3commRevocationStatusV1), status["type"])
		assert.Equal(t, authCredentialExpiration, response2.Vc.Expiration.UTC().Unix())
	})
}
