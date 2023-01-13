package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_CreateIdentity(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	claimsService := services.NewClaim(cfg.ReverseHashService.Enabled, cfg.ReverseHashService.URL, cfg.ServerUrl, claimsRepo, storage, mtService)
	schemaService := services.NewSchema(storage)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	claimsService := services.NewClaim(cfg.ReverseHashService.Enabled, cfg.ReverseHashService.URL, cfg.ServerUrl, claimsRepo, schemaService, identityService, storage)

	server := NewServer(&cfg, identityService, claimsService, schemaService)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "should create an identity",
			expected: expected{
				httpCode: 201,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/identities", nil)
			handler.ServeHTTP(rr, req)

			var response CreateIdentityResponse
			assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			assert.NotNil(t, *response.State.ClaimsTreeRoot)
			assert.NotNil(t, response.State.CreatedAt)
			assert.NotNil(t, response.State.ModifiedAt)
			assert.NotNil(t, response.State.State)
			assert.NotNil(t, response.State.Status)
			assert.NotNil(t, *response.Identifier)
			assert.NotNil(t, response.Immutable)
		})
	}
}

func TestServer_RevokeClaim(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	claimsService := services.NewClaim(cfg.ReverseHashService.Enabled, cfg.ReverseHashService.URL, cfg.ServerUrl, claimsRepo, storage, mtService)
	schemaService := services.NewSchema(storage)
	server := NewServer(&cfg, identityService, claimsService, schemaService)

	idStr := "did:polygonid:polygon:mumbai:2qM77fA6NGGWL9QEeb1dv2VA6wz5svcohgv61LZ7wB"
	identity := &domain.Identity{
		Identifier: idStr,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	idClaim, _ := uuid.NewUUID()
	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)
	fixture.CreateClaim(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      &idStr,
		Issuer:          idStr,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "",
		Expiration:      0,
		Version:         0,
		RevNonce:        revNonce,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	query := tests.ExecQueryParams{
		Query: `INSERT INTO identity_mts (identifier, type) VALUES 
                                                    ($1, 0),
                                                    ($1, 1),
                                                    ($1, 2),
                                                    ($1, 3)`,
		Arguments: []interface{}{idStr},
	}

	fixture.ExecQuery(t, query)

	handler := getHandler(context.Background(), server)

	type expected struct {
		response RevokeClaimResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		did      string
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "should revoke the claim",
			did:   idStr,
			nonce: nonce,
			expected: expected{
				httpCode: 202,
				response: RevokeClaim202JSONResponse{
					Status: "pending",
				},
			},
		},
		{
			name:  "should get an error wrong nonce",
			did:   idStr,
			nonce: int64(1231323),
			expected: expected{
				httpCode: 404,
				response: RevokeClaim404JSONResponse{N404JSONResponse{
					Message: "the claim does not exist",
				}},
			},
		},
		{
			name:  "should get an error",
			did:   "did:polygonid:polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			nonce: nonce,
			expected: expected{
				httpCode: 500,
				response: RevokeClaim500JSONResponse{N500JSONResponse{
					Message: "error gettting merkles trees: not found",
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims/revoke/%d", tc.did, tc.nonce)
			req, _ := http.NewRequest("POST", url, nil)
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case RevokeClaim202JSONResponse:
				var response RevokeClaim202JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Status, v.Status)
			case RevokeClaim404JSONResponse:
				var response RevokeClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case RevokeClaim500JSONResponse:
				var response RevokeClaim500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			default:
				t.Fail()
			}
		})
	}
}
