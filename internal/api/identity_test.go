package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	core "github.com/iden3/go-iden3-core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: string(core.Amoy), Type: BJJ},
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: network, Type: BJJ},
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: blockchain, Method: method, Network: "mynetwork", Type: BJJ},
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: blockchain, Method: "my method", Network: network, Type: BJJ},
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: BJJ},
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
					AuthBJJCredentialStatus *CreateIdentityRequestDidMetadataAuthBJJCredentialStatus `json:"authBJJCredentialStatus,omitempty"`
					Blockchain              string                                                   `json:"blockchain"`
					Method                  string                                                   `json:"method"`
					Network                 string                                                   `json:"network"`
					Type                    CreateIdentityRequestDidMetadataType                     `json:"type"`
				}{Blockchain: "my blockchain", Method: method, Network: network, Type: "a wrong type"},
			},
			expected: expected{
				httpCode: 400,
				message:  common.ToPointer("Type must be BJJ or ETH"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/v1/identities", tests.JSONBody(t, tc.input))
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
}

func TestServer_GetIdentities(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	identity1 := &domain.Identity{Identifier: "did:polygonid:polygon:mumbai:2qE1ZT16aqEWhh9mX9aqM2pe2ZwV995dTkReeKwCaQ"}
	identity2 := &domain.Identity{Identifier: "did:polygonid:polygon:mumbai:2qMHFTHn2SC3XkBEJrR4eH4Yk8jRGg5bzYYG1ZGECa"}
	fixture := tests.NewFixture(storage)
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
			req, err := http.NewRequest("GET", "/v1/identities", nil)
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
