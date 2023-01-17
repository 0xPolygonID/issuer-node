package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_CreateIdentity(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, storage, claimsConf)

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
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, storage, claimsConf)

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

func TestServer_GetIdentities(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, storage, claimsConf)
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
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, storage, claimsConf)

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

	type credentialKYCSubject struct {
		Id           string `json:"id"`
		Birthday     uint64 `json:"birthday"`
		DocumentType uint64 `json:"documentType"`
		Type         string `json:"type"`
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
				var response GetClaim200JSONResponse
				var responseCredentialStatus verifiable.CredentialStatus
				var responseCredentialSubject, tcCredentialSubject credentialKYCSubject
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Id, v.Id)
				assert.Equal(t, len(response.Context), len(v.Context))
				assert.EqualValues(t, response.Context, v.Context)
				assert.EqualValues(t, response.CredentialSchema, v.CredentialSchema)
				assert.NoError(t, mapstructure.Decode(response.CredentialSubject, &responseCredentialSubject))
				assert.NoError(t, mapstructure.Decode(v.CredentialSubject, &tcCredentialSubject))
				assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
				assert.InDelta(t, response.IssuanceDate.Unix(), v.IssuanceDate.Unix(), 30)
				assert.Equal(t, response.Type, v.Type)
				assert.NoError(t, mapstructure.Decode(response.CredentialStatus, &responseCredentialStatus))
				credentialStatusTC, ok := v.CredentialStatus.(verifiable.CredentialStatus)
				require.True(t, ok)
				assert.EqualValues(t, responseCredentialStatus, credentialStatusTC)
				assert.Equal(t, response.Expiration, v.Expiration)
				assert.Equal(t, response.Issuer, v.Issuer)

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
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	schemaService := services.NewSchema(storage)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, storage, claimsConf)
	fixture := tests.NewFixture(storage)
	server := NewServer(&cfg, identityService, claimsService, schemaService)

	ctx := context.Background()
	identity, err := server.identityService.Create(ctx, "https://localhost.com")
	require.NoError(t, err)

	defaultClaim := fixture.GetDefaultAuthClaimOfIssuer(t, identity.Identifier)
	defaultClaimVC, err := server.schemaService.FromClaimModelToW3CCredential(*defaultClaim)
	assert.NoError(t, err)

	emtpyIdentityStr := "did:polygonid:polygon:mumbai:2qLQGgjpP5Yq7r7jbRrQZbWy8ikADvxamSLB7CqR4F"

	handler := getHandler(context.Background(), server)

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

	credentialSubjectTypes := []string{"AuthBJJCredential", "KYCAgeCredential"}

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
			did:  emtpyIdentityStr,
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
				var responseCredentialStatus verifiable.CredentialStatus
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, len(response), tc.expected.len)
				for i := range response {
					assert.Equal(t, response[i].Id, v[i].Id)
					assert.Equal(t, len(response[i].Context), len(v[i].Context))
					assert.EqualValues(t, response[i].Context, v[i].Context)
					assert.EqualValues(t, response[i].CredentialSchema, v[i].CredentialSchema)
					assert.InDelta(t, response[i].IssuanceDate.Unix(), v[i].IssuanceDate.Unix(), 30)
					assert.Equal(t, response[i].Type, v[i].Type)
					assert.Equal(t, response[i].Expiration, v[i].Expiration)
					assert.Equal(t, response[i].Issuer, v[i].Issuer)
					credentialSubjectType, ok := v[i].CredentialSubject["type"]
					require.True(t, ok)
					assert.Contains(t, credentialSubjectTypes, credentialSubjectType)
					if credentialSubjectType == "AuthBJJCredential" {
						var responseCredentialSubject, tcCredentialSubject credentialBJJSubject
						assert.NoError(t, mapstructure.Decode(response[i].CredentialSubject, &responseCredentialSubject))
						assert.NoError(t, mapstructure.Decode(v[i].CredentialSubject, &tcCredentialSubject))
						assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
					} else {
						var responseCredentialSubject, tcCredentialSubject credentialKYCSubject
						assert.NoError(t, mapstructure.Decode(response[i].CredentialSubject, &responseCredentialSubject))
						assert.NoError(t, mapstructure.Decode(v[i].CredentialSubject, &tcCredentialSubject))
						assert.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
					}

					assert.NoError(t, mapstructure.Decode(response[i].CredentialStatus, &responseCredentialStatus))
					responseCredentialStatus.ID = strings.Replace(responseCredentialStatus.ID, "%3A", ":", -1)
					credentialStatusTC, ok := v[i].CredentialStatus.(verifiable.CredentialStatus)
					require.True(t, ok)
					assert.EqualValues(t, responseCredentialStatus, credentialStatusTC)
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
