package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_CreateVerification(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  *string
	}
	type testConfig struct {
		name       string
		auth       func() (string, string)
		identifier string
		body       CreateVerificationQueryRequest
		expected   expected
	}

	idFromString, err := core.IDFromString("x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp")
	require.NoError(t, err)

	did, err := core.ParseDIDFromID(idFromString)
	require.NoError(t, err)

	identity := &domain.Identity{
		Identifier: did.String(),
	}
	fixture := repositories.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	for _, tc := range []testConfig{
		{
			name:       "Valid request",
			auth:       authOk,
			identifier: did.String(),
			body: CreateVerificationQueryRequest{
				ChainId:             137,
				SkipRevocationCheck: false,
				Scopes: []map[string]interface{}{
					{
						"circuitId": "credentialAtomicQuerySigV2",
						"id":        1,
						"params": map[string]interface{}{
							"nullifierSessionID": "123456789",
						},
						"query": nil,
					},
				},
			},
			expected: expected{
				httpCode: http.StatusCreated,
			},
		},
		{
			name:       "Invalid identifier",
			auth:       authOk,
			identifier: "invalid-identifier",
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("invalid issuer did"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/verification", tc.identifier)
			req, err := http.NewRequest("POST", url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusCreated {
				var response CreateVerificationQueryResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.VerificationQueryId)
			} else if tc.expected.httpCode == http.StatusBadRequest {
				var response CreateVerification400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_CheckVerification(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	idFromString, err := core.IDFromString("x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp")
	require.NoError(t, err)

	did, err := core.ParseDIDFromID(idFromString)
	require.NoError(t, err)

	// Mock data setup
	queryWithoutResponse := domain.VerificationQuery{
		ID:                  uuid.New(),
		IssuerDID:           did.String(),
		ChainID:             137,
		SkipCheckRevocation: false,
		// CreatedAt:           time.Now(),
	}
	queryWithResponse := domain.VerificationQuery{
		ID:                  uuid.New(),
		IssuerDID:           did.String(),
		ChainID:             137,
		SkipCheckRevocation: false,
		// CreatedAt:           time.Now(),
	}

	responseJson := pgtype.JSONB{
		Bytes:  []byte(`{"foo":"bar"}`),
		Status: pgtype.Present,
	}

	response := domain.VerificationResponse{
		ID:                  uuid.New(),
		VerificationQueryID: queryWithResponse.ID,
		UserDID:             did.String(),
		Response:            &responseJson,
		Pass:                true,
		// CreatedAt:           time.Now(),
	}

	fixture := repositories.NewFixture(storage)
	fixture.CreateVerificationQuery(t, *did, queryWithoutResponse)
	fixture.CreateVerificationQuery(t, *did, queryWithResponse)
	fixture.CreateVerificationResponse(t, queryWithResponse.ID, response)

	type expected struct {
		httpCode     int
		responseType string
	}
	type testConfig struct {
		name       string
		auth       func() (string, string)
		identifier string
		queryID    string
		expected   expected
	}

	for _, tc := range []testConfig{
		{
			name:       "Verification response exists",
			auth:       authOk,
			identifier: did.String(),
			queryID:    queryWithResponse.ID.String(),
			expected: expected{
				httpCode:     http.StatusOK,
				responseType: "VerificationResponse",
			},
		},
		{
			name:       "Verification query exists without response",
			auth:       authOk,
			identifier: did.String(),
			queryID:    queryWithoutResponse.ID.String(),
			expected: expected{
				httpCode:     http.StatusOK,
				responseType: "VerificationQueryRequest",
			},
		},
		{
			name:       "Query not found",
			auth:       authOk,
			identifier: did.String(),
			queryID:    uuid.New().String(),
			expected: expected{
				httpCode: http.StatusNotFound,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/verification?id=%s", tc.identifier, tc.queryID)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)
			if tc.expected.httpCode == http.StatusOK {
				var response CheckVerificationResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				if tc.expected.responseType == "VerificationResponse" {
					assert.NotNil(t, response.VerificationResponse)
					assert.Nil(t, response.VerificationQueryRequest)
				} else if tc.expected.responseType == "VerificationQueryRequest" {
					assert.NotNil(t, response.VerificationQueryRequest)
					assert.Nil(t, response.VerificationResponse)
				}
			}
		})
	}
}
