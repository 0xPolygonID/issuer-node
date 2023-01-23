package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func TestServer_CreateIdentity(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

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
			req, err := http.NewRequest("POST", "/v1/identities", nil)
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			var response CreateIdentityResponse
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
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
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaService)

	idStr := "did:polygonid:polygon:mumbai:2qM77fA6NGGWL9QEeb1dv2VA6wz5svcohgv61LZ7wB"
	identity := &domain.Identity{
		Identifier: idStr,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	idClaim, err := uuid.NewUUID()
	require.NoError(t, err)
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
					Message: "error getting merkle trees: not found",
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims/revoke/%d", tc.did, tc.nonce)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			require.NoError(t, err)
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
				require.Fail(t, "unexpected http response", tc.expected.httpCode)
			}
		})
	}
}

func TestServer_CreateClaim(t *testing.T) {
	ctx := log.NewContext(context.Background(), log.LevelDebug, log.OutputText, os.Stdout)

	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaService)
	handler := getHandler(ctx, server)

	iden, err := identityService.Create(ctx, "polygon-test")
	require.NoError(t, err)
	did := iden.Identifier

	type expected struct {
		response CreateClaimResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		did      string
		body     CreateClaimRequest
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Happy path",
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(int64(12345)),
			},
			expected: expected{
				response: CreateClaim201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Wrong credential url",
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "wrong url",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(int64(12345)),
			},
			expected: expected{
				response: CreateClaim400JSONResponse{N400JSONResponse{Message: "malformed url"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Unreachable well formed credential url",
			did:  did,
			body: CreateClaimRequest{
				CredentialSchema: "http://www.wrong.url/cannot/get/the/credential",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(int64(12345)),
			},
			expected: expected{
				response: CreateClaim422JSONResponse{N422JSONResponse{Message: "cannot load schema"}},
				httpCode: http.StatusUnprocessableEntity,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims", tc.did)

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateClaimResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response CreateClaim400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusUnprocessableEntity:
				var response CreateClaim422JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			default:
				require.Fail(t, "unexpected http status response", tc.expected.httpCode)
			}
		})
	}
}

func TestServer_GetIdentities(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)
	server := NewServer(&cfg, identityService, claimsService, schemaService)
	handler := getHandler(context.Background(), server)

	idStr1 := "did:polygonid:polygon:mumbai:2qE1ZT16aqEWhh9mX9aqM2pe2ZwV995dTkReeKwCaQ"
	idStr2 := "did:polygonid:polygon:mumbai:2qMHFTHn2SC3XkBEJrR4eH4Yk8jRGg5bzYYG1ZGECa"
	identity1 := &domain.Identity{
		Identifier: idStr1,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	identity2 := &domain.Identity{
		Identifier: idStr2,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity1)
	fixture.CreateIdentity(t, identity2)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "should return all the entities",
			expected: expected{
				httpCode: 200,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/v1/identities", nil)
			handler.ServeHTTP(rr, req)

			var response GetIdentities200JSONResponse
			assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			assert.Equal(t, tc.expected.httpCode, rr.Code)
			assert.True(t, len(response) >= 2)
		})
	}
}

func TestServer_GetClaim(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaService)

	idStr := "did:polygonid:polygon:mumbai:2qLduMv2z7hnuhzkcTWesCUuJKpRVDEThztM4tsJUj"
	idStrWithoutClaims := "did:polygonid:polygon:mumbai:2qGjTUuxZKqKS4Q8UmxHUPw55g15QgEVGnj6Wkq8Vk"
	identity := &domain.Identity{
		Identifier: idStr,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

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
		response GetClaimResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		did      string
		claimID  uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "should get an error non existing claimID",
			did:     idStr,
			claimID: uuid.New(),
			expected: expected{
				httpCode: 404,
				response: GetClaim404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
			},
		},
		{
			name:    "should get an error the given did has no entry for claimID",
			did:     idStrWithoutClaims,
			claimID: claim.ID,
			expected: expected{
				httpCode: 404,
				response: GetClaim404JSONResponse{N404JSONResponse{
					Message: "claim not found",
				}},
			},
		},
		{
			name:    "should get an error wrong did invalid format",
			did:     ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			claimID: claim.ID,
			expected: expected{
				httpCode: 400,
				response: GetClaim400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
			},
		},
		{
			name:    "should get the claim",
			did:     idStr,
			claimID: claim.ID,
			expected: expected{
				httpCode: 200,
				response: GetClaim200JSONResponse{
					Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
					CredentialSchema: CredentialSchema{
						"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
						"JsonSchemaValidator2018",
					},
					CredentialStatus: verifiable.CredentialStatus{
						ID:              fmt.Sprintf("http://localhost/api/v1/identities/%s/claims/revocation/status/%d", idStr, claim.RevNonce),
						Type:            "SparseMerkleTreeProof",
						RevocationNonce: uint64(claim.RevNonce),
					},
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     float64(19960424),
						"documentType": float64(2),
						"type":         "KYCAgeCredential",
					},
					Id:           fmt.Sprintf("http://localhost/api/v1/claim/%s", claim.ID),
					IssuanceDate: common.ToPointer(time.Now().UTC()),
					Issuer:       idStr,
					Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims/%s", tc.did, tc.claimID.String())
			req, _ := http.NewRequest("GET", url, nil)
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetClaim200JSONResponse:
				var response GetClaimResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				validateClaim(t, response, GetClaimResponse(v))

			case GetClaim400JSONResponse:
				var response GetClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetClaim404JSONResponse:
				var response GetClaim404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetClaim500JSONResponse:
				var response GetClaim500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			default:
				t.Fail()
			}
		})
	}
}

func TestServer_GetClaims(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	fixture := tests.NewFixture(storage)
	server := NewServer(&cfg, identityService, claimsService, schemaService)

	ctx := context.Background()
	identity, err := server.identityService.Create(ctx, "https://localhost.com")
	require.NoError(t, err)

	defaultClaim := fixture.GetDefaultAuthClaimOfIssuer(t, identity.Identifier)
	defaultClaimVC, err := server.schemaService.FromClaimModelToW3CCredential(*defaultClaim)
	assert.NoError(t, err)

	identityMultipleClaims, err := server.identityService.Create(ctx, "https://localhost.com")
	require.NoError(t, err)

	defaultClaimMultipleClaims := fixture.GetDefaultAuthClaimOfIssuer(t, identityMultipleClaims.Identifier)
	defaultClaimMultipleClaimsVC, err := server.schemaService.FromClaimModelToW3CCredential(*defaultClaimMultipleClaims)
	assert.NoError(t, err)

	claim := fixture.NewClaim(t, defaultClaimMultipleClaimsVC.Issuer)
	_ = fixture.CreateClaim(t, claim)

	emptyIdentityStr := "did:polygonid:polygon:mumbai:2qLQGgjpP5Yq7r7jbRrQZbWy8ikADvxamSLB7CqR4F"

	handler := getHandler(context.Background(), server)

	type expected struct {
		response GetClaimsResponseObject
		len      int
		httpCode int
	}

	type testConfig struct {
		name     string
		did      string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "should get an error wrong did invalid format",
			did:  ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetClaims400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
			},
		},
		{
			name: "should get 0 claims",
			did:  emptyIdentityStr,
			expected: expected{
				httpCode: http.StatusOK,
				len:      0,
				response: GetClaims200JSONResponse{},
			},
		},
		{
			name: "should get the default claim for a did that has no created claims",
			did:  identity.Identifier,
			expected: expected{
				httpCode: http.StatusOK,
				len:      1,
				response: GetClaims200JSONResponse{
					GetClaimResponse{
						Id:      defaultClaimVC.ID,
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://schema.iden3.io/core/jsonld/auth.jsonld"},
						CredentialSchema: CredentialSchema{
							"https://schema.iden3.io/core/json/auth.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("https://localhost.com/api/v1/identities/%s/claims/revocation/status/%d", identity.Identifier, 0),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: 0,
						},
						CredentialSubject: map[string]interface{}{
							"type": "AuthBJJCredential",
							"x":    defaultClaimVC.CredentialSubject["x"],
							"y":    defaultClaimVC.CredentialSubject["y"],
						},
						IssuanceDate: common.ToPointer(time.Now().UTC()),
						Issuer:       identity.Identifier,
						Type:         []string{"VerifiableCredential", "AuthBJJCredential"},
					},
				},
			},
		},
		{
			name: "should get the default claim plus another one that has been created",
			did:  identityMultipleClaims.Identifier,
			expected: expected{
				httpCode: http.StatusOK,
				len:      2,
				response: GetClaims200JSONResponse{
					GetClaimResponse{
						Id:      defaultClaimMultipleClaimsVC.ID,
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://schema.iden3.io/core/jsonld/auth.jsonld"},
						CredentialSchema: CredentialSchema{
							"https://schema.iden3.io/core/json/auth.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("https://localhost.com/api/v1/identities/%s/claims/revocation/status/%d", identityMultipleClaims.Identifier, 0),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: 0,
						},
						CredentialSubject: map[string]interface{}{
							"type": "AuthBJJCredential",
							"x":    defaultClaimMultipleClaimsVC.CredentialSubject["x"],
							"y":    defaultClaimMultipleClaimsVC.CredentialSubject["y"],
						},
						IssuanceDate: common.ToPointer(time.Now().UTC()),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "AuthBJJCredential"},
					},
					GetClaimResponse{
						Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
						CredentialSchema: CredentialSchema{
							"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
							"JsonSchemaValidator2018",
						},
						CredentialStatus: verifiable.CredentialStatus{
							ID:              fmt.Sprintf("http://localhost/api/v1/identities/%s/claims/revocation/status/%d", identityMultipleClaims.Identifier, claim.RevNonce),
							Type:            "SparseMerkleTreeProof",
							RevocationNonce: uint64(claim.RevNonce),
						},
						CredentialSubject: map[string]interface{}{
							"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
							"birthday":     float64(19960424),
							"documentType": float64(2),
							"type":         "KYCAgeCredential",
						},
						Id:           fmt.Sprintf("http://localhost/api/v1/claim/%s", claim.ID),
						IssuanceDate: common.ToPointer(time.Now().UTC()),
						Issuer:       identityMultipleClaims.Identifier,
						Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims", tc.did)
			req, _ := http.NewRequest("GET", url, nil)
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetClaims200JSONResponse:
				var response GetClaims200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, len(response), tc.expected.len)
				for i := range response {
					validateClaim(t, response[i], v[i])
				}
			case GetClaims400JSONResponse:
				var response GetClaims400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case GetClaims500JSONResponse:
				var response GetClaims500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			default:
				t.Fail()
			}
		})
	}
}

func TestServer_GetRevocationStatus(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(storage)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "https://host.com",
	}

	identity, err := identityService.Create(ctx, "http://localhost:3001")
	assert.NoError(t, err)
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)
	server := NewServer(&cfg, identityService, claimsService, schemaService)
	handler := getHandler(context.Background(), server)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, _ := core.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	expiration := int64(12345)

	merklizedRootPosition := "value"
	claim, err := claimsService.CreateClaim(context.Background(), ports.NewCreateClaimRequest(did, schema, credentialSubject, &expiration, typeC, nil, nil, &merklizedRootPosition))
	assert.NoError(t, err)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "should get revocation status",
			nonce: int64(claim.RevNonce),
			expected: expected{
				httpCode: 200,
			},
		},

		{
			name:  "should get revocation status wrong nonce",
			nonce: 123456,
			expected: expected{
				httpCode: 200,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims/revocation/status/%d", identity.Identifier, tc.nonce)
			req, _ := http.NewRequest("GET", url, nil)
			handler.ServeHTTP(rr, req)

			var response GetRevocationStatus200JSONResponse
			assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			assert.Equal(t, tc.expected.httpCode, rr.Code)
			assert.NotNil(t, response.Issuer.ClaimsTreeRoot)
			assert.NotNil(t, response.Issuer.State)
			assert.NotNil(t, response.Mtp.Existence)
			assert.NotNil(t, response.Mtp.Siblings)
		})
	}
}

func validateClaim(t *testing.T, resp, tc GetClaimResponse) {
	var responseCredentialStatus verifiable.CredentialStatus

	credentialSubjectTypes := []string{"AuthBJJCredential", "KYCAgeCredential"}

	type credentialKYCSubject struct {
		Id           string `json:"id"`
		Birthday     uint64 `json:"birthday"`
		DocumentType uint64 `json:"documentType"`
		Type         string `json:"type"`
	}

	type credentialBJJSubject struct {
		Type string `json:"type"`
		X    string `json:"x"`
		Y    string `json:"y"`
	}

	assert.Equal(t, resp.Id, tc.Id)
	assert.Equal(t, len(resp.Context), len(tc.Context))
	assert.EqualValues(t, resp.Context, tc.Context)
	assert.EqualValues(t, resp.CredentialSchema, tc.CredentialSchema)
	assert.InDelta(t, resp.IssuanceDate.Unix(), tc.IssuanceDate.Unix(), 30)
	assert.Equal(t, resp.Type, tc.Type)
	assert.Equal(t, resp.Expiration, tc.Expiration)
	assert.Equal(t, resp.Issuer, tc.Issuer)
	credentialSubjectType, ok := tc.CredentialSubject["type"]
	require.True(t, ok)
	assert.Contains(t, credentialSubjectTypes, credentialSubjectType)
	if credentialSubjectType == "AuthBJJCredential" {
		var responseCredentialSubject, tcCredentialSubject credentialBJJSubject
		assert.NoError(t, mapstructure.Decode(resp.CredentialSubject, &responseCredentialSubject))
		assert.NoError(t, mapstructure.Decode(tc.CredentialSubject, &tcCredentialSubject))
		assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
	} else {
		var responseCredentialSubject, tcCredentialSubject credentialKYCSubject
		assert.NoError(t, mapstructure.Decode(resp.CredentialSubject, &responseCredentialSubject))
		assert.NoError(t, mapstructure.Decode(tc.CredentialSubject, &tcCredentialSubject))
		assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
	}

	assert.NoError(t, mapstructure.Decode(resp.CredentialStatus, &responseCredentialStatus))
	responseCredentialStatus.ID = strings.Replace(responseCredentialStatus.ID, "%3A", ":", -1)
	credentialStatusTC, ok := tc.CredentialStatus.(verifiable.CredentialStatus)
	require.True(t, ok)
	assert.EqualValues(t, responseCredentialStatus, credentialStatusTC)
}
