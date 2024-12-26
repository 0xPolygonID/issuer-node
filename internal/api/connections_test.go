package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_CreateConnection(t *testing.T) {
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

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	const serviceContext = "https://www.w3.org/ns/did/v1"
	userDidDoc := verifiable.DIDDocument{
		Context: []string{serviceContext},
		ID:      userDID.ID,
		Service: []interface{}{},
	}
	var userDoc map[string]interface{}
	userDocBytes, err := json.Marshal(userDidDoc)
	require.NoError(t, err)
	err = json.Unmarshal(userDocBytes, &userDoc)
	require.NoError(t, err)

	issuerDidDoc := verifiable.DIDDocument{
		Context: []string{serviceContext},
		ID:      userDID.ID,
		Service: []interface{}{},
	}
	var issuerDoc map[string]interface{}
	issuerDocBytes, err := json.Marshal(issuerDidDoc)
	require.NoError(t, err)
	err = json.Unmarshal(issuerDocBytes, &issuerDoc)
	require.NoError(t, err)

	wrongDidDoc := make(map[string]interface{})
	wrongDidDoc["wrong"] = "wrong"

	type expected struct {
		httpCode int
		message  *string
	}

	type testConfig struct {
		name      string
		connID    uuid.UUID
		auth      func() (string, string)
		body      *CreateConnectionRequest
		issuerDID string
		expected  expected
	}

	for _, tc := range []testConfig{
		{
			name:      "No auth header",
			auth:      authWrong,
			issuerDID: issuerDID.String(),
			body: &CreateConnectionRequest{
				UserDID:   userDID.String(),
				UserDoc:   userDoc,
				IssuerDoc: issuerDoc,
			},
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "should get an error, wrong user did document",
			connID:    uuid.New(),
			auth:      authOk,
			issuerDID: issuerDID.String(),
			body: &CreateConnectionRequest{
				UserDID:   userDID.String(),
				UserDoc:   wrongDidDoc,
				IssuerDoc: issuerDoc,
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("invalid user did document"),
			},
		},
		{
			name:      "should get an error, wrong issuer did document",
			connID:    uuid.New(),
			auth:      authOk,
			issuerDID: issuerDID.String(),
			body: &CreateConnectionRequest{
				UserDID:   userDID.String(),
				UserDoc:   userDoc,
				IssuerDoc: wrongDidDoc,
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("invalid issuer did document"),
			},
		},
		{
			name:      "should get an error, wrong user did",
			connID:    uuid.New(),
			auth:      authOk,
			issuerDID: issuerDID.String(),
			body: &CreateConnectionRequest{
				UserDID:   "invalid did",
				UserDoc:   userDoc,
				IssuerDoc: issuerDoc,
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("invalid user did"),
			},
		},
		{
			name:      "should get an error, wrong issuer did",
			connID:    uuid.New(),
			auth:      authOk,
			issuerDID: "invalid did",
			body: &CreateConnectionRequest{
				UserDID:   userDID.String(),
				UserDoc:   userDoc,
				IssuerDoc: issuerDoc,
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("invalid issuer did"),
			},
		},
		{
			name:      "connetion created",
			connID:    uuid.New(),
			auth:      authOk,
			issuerDID: issuerDID.String(),
			body: &CreateConnectionRequest{
				UserDID:   userDID.String(),
				UserDoc:   userDoc,
				IssuerDoc: issuerDoc,
			},
			expected: expected{
				httpCode: http.StatusCreated,
				message:  common.ToPointer(""),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			urlTest := fmt.Sprintf("/v2/identities/%s/connections", tc.issuerDID)
			parsedURL, err := url.Parse(urlTest)
			require.NoError(t, err)

			reqBytes, err := json.Marshal(tc.body)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, parsedURL.String(), bytes.NewBuffer(reqBytes))
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response DeleteConnection400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			case http.StatusOK:
				var response DeleteConnection200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_DeleteConnection(t *testing.T) {
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

	fixture := repositories.NewFixture(storage)

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	userDID2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R64")
	require.NoError(t, err)

	conn := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	conn2 := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID2,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	_ = fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID2.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        1,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	type expected struct {
		httpCode int
		message  *string
	}

	type testConfig struct {
		name             string
		connID           uuid.UUID
		deleteCredential bool
		revokeCredential bool
		auth             func() (string, string)
		expected         expected
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
			name:   "should get an error, not existing connection",
			connID: uuid.New(),
			auth:   authOk,
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("The given connection does not exist"),
			},
		},
		{
			name:   "should delete the connection",
			connID: conn,
			auth:   authOk,
			expected: expected{
				httpCode: http.StatusOK,
				message:  common.ToPointer("Connection successfully deleted."),
			},
		},
		{
			name:             "should delete the connection and revoke + delete credentials",
			connID:           conn2,
			deleteCredential: true,
			revokeCredential: true,
			auth:             authOk,
			expected: expected{
				httpCode: http.StatusOK,
				message:  common.ToPointer("Connection successfully deleted. Credentials successfully deleted. Credentials successfully revoked."),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			urlTest := fmt.Sprintf("/v2/identities/%s/connections/%s", issuerDID, tc.connID.String())
			parsedURL, err := url.Parse(urlTest)
			require.NoError(t, err)
			values := parsedURL.Query()
			if tc.deleteCredential {
				values.Add("deleteCredentials", "true")
			}
			if tc.revokeCredential {
				values.Add("revokeCredentials", "true")
			}
			parsedURL.RawQuery = values.Encode()
			req, err := http.NewRequest(http.MethodDelete, parsedURL.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response DeleteConnection400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			case http.StatusOK:
				var response DeleteConnection200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_DeleteConnectionCredentials(t *testing.T) {
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	fixture := repositories.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	conn := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	_ = fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246563",
	})

	_ = fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246562",
	})

	type expected struct {
		httpCode int
		message  *string
	}

	type testConfig struct {
		name     string
		connID   uuid.UUID
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
			name:   "should delete the connection",
			connID: conn,
			auth:   authOk,
			expected: expected{
				httpCode: http.StatusOK,
				message:  common.ToPointer("Credentials of the connection successfully deleted"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/connections/%s/credentials", issuerDID, tc.connID.String())
			req, err := http.NewRequest("DELETE", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response DeleteConnectionCredentials200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_RevokeConnectionCredentials(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	iden, err := server.Services.identity.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	fixture := repositories.NewFixture(storage)

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	conn := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	_ = fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        1,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	type expected struct {
		httpCode int
		message  *string
	}

	type testConfig struct {
		name     string
		connID   uuid.UUID
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
			name:   "should revoke the connection credentials",
			connID: conn,
			auth:   authOk,
			expected: expected{
				httpCode: http.StatusAccepted,
				message:  common.ToPointer("Credentials revocation request sent"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/connections/%s/credentials/revoke", issuerDID, tc.connID.String())
			req, err := http.NewRequest("POST", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusAccepted:
				var response RevokeConnectionCredentials202JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_GetConnectionsDefaultSort(t *testing.T) {
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
	fixture := repositories.NewFixture(storage)
	type expected struct {
		httpCode int
		response GetConnectionsResponseObject
	}

	type testConfig struct {
		name       string
		issuerDID  string
		auth       func() (string, string)
		page       int
		maxResults int
		expected   expected
	}

	expectedConnections := createConnections(t, issuerDID, fixture)

	for _, tc := range []testConfig{
		{
			name:      "No auth header",
			auth:      authWrong,
			issuerDID: issuerDID.String(),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:       "Wrong issuer did",
			auth:       authOk,
			page:       1,
			maxResults: 10,
			issuerDID:  "wrong did",
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetConnections400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "invalid issuer did",
					},
				},
			},
		},
		{
			name:       "Wrong page",
			auth:       authOk,
			page:       0,
			maxResults: 10,
			issuerDID:  issuerDID.String(),
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetConnections400JSONResponse{
					N400JSONResponse: N400JSONResponse{
						Message: "page must be greater than 0",
					},
				},
			},
		},
		{
			name:       "should return 10 connection",
			auth:       authOk,
			page:       1,
			maxResults: 10,
			issuerDID:  issuerDID.String(),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Meta: PaginatedMetadata{
						MaxResults: 10,
						Total:      12,
						Page:       1,
					},
					Items: expectedConnections[:10],
				},
			},
		},
		{
			name:       "should return 2 connection",
			auth:       authOk,
			page:       2,
			maxResults: 10,
			issuerDID:  issuerDID.String(),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Meta: PaginatedMetadata{
						MaxResults: 10,
						Total:      12,
						Page:       2,
					},
					Items: expectedConnections[10:12],
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			urlTest := fmt.Sprintf("/v2/identities/%s/connections?max_results=%d&page=%d", tc.issuerDID, tc.maxResults, tc.page)
			parsedURL, err := url.Parse(urlTest)
			require.NoError(t, err)

			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response GetConnections400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response, response)
			case http.StatusOK:
				var response GetConnections200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expectedResponse, ok := tc.expected.response.(GetConnections200JSONResponse)
				require.True(t, ok)
				assert.Equal(t, len(expectedResponse.Items), len(response.Items))
				for i := range response.Items {
					assert.Equal(t, expectedResponse.Items[i].Id, response.Items[i].Id)
					assert.Equal(t, expectedResponse.Items[i].IssuerID, response.Items[i].IssuerID)
					assert.Equal(t, expectedResponse.Items[i].UserID, response.Items[i].UserID)
				}
				assert.Equal(t, int(expectedResponse.Meta.MaxResults), int(response.Meta.MaxResults))
				assert.Equal(t, int(expectedResponse.Meta.Total), int(response.Meta.Total))
				assert.Equal(t, expectedResponse.Meta.Page, response.Meta.Page)
			}
		})
	}
}

func createConnections(t *testing.T, issuerDID *w3c.DID, fixture *repositories.Fixture) []GetConnectionResponse {
	t.Helper()
	usersDIDs := []string{
		"did:polygonid:polygon:amoy:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R64",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R65",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R66",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R67",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R68",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R69",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R70",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R71",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R72",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R73",
		"did:polygonid:polygon:amoy:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R74",
	}
	connections := make([]GetConnectionResponse, len(usersDIDs))
	for i, userDID := range usersDIDs {
		user, err := w3c.ParseDID(userDID)
		require.NoError(t, err)
		conn := fixture.CreateConnection(t, &domain.Connection{
			ID:         uuid.New(),
			IssuerDID:  *issuerDID,
			UserDID:    *user,
			IssuerDoc:  nil,
			UserDoc:    nil,
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		})
		connections[len(usersDIDs)-i-1] = GetConnectionResponse{
			Id:        conn.String(),
			IssuerID:  issuerDID.String(),
			UserID:    user.String(),
			CreatedAt: TimeUTC(time.Now()),
		}
	}
	return connections
}
