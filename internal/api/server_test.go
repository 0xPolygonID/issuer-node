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
	claimsService := services.NewClaim(cfg.ReverseHashService.Enabled, cfg.ReverseHashService.URL, cfg.ServerUrl, claimsRepo, schemaService, identityService, mtService, storage)

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

	claimID, _ := uuid.NewUUID()
	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)

	claim := &domain.Claim{
		ID:              claimID,
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
	}
	vc := verifiable.W3CCredential{
		ID:           fmt.Sprintf("http://localhost/api/v1/claim/%s", claimID),
		Context:      []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
		Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
		IssuanceDate: common.ToPointer(time.Now().UTC()),
		CredentialSubject: map[string]interface{}{
			"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
			"birthday":     19960424,
			"documentType": 2,
			"type":         "KYCAgeCredential",
		},
		CredentialStatus: verifiable.CredentialStatus{
			ID:              fmt.Sprintf("http://localhost/api/v1/identities/%s/claims/revocation/status/%d", idStr, revNonce),
			Type:            "SparseMerkleTreeProof",
			RevocationNonce: uint64(nonce),
		},
		Issuer: idStr,
		CredentialSchema: verifiable.CredentialSchema{
			ID:   "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
			Type: "JsonSchemaValidator2018",
		},
	}

	require.NoError(t, claim.CredentialStatus.Set(vc.CredentialStatus))
	require.NoError(t, claim.Data.Set(vc))

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

	type credentialSubject struct {
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
			claimID: claimID,
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
			claimID: claimID,
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
			claimID: claimID,
			expected: expected{
				httpCode: 200,
				response: GetClaim200JSONResponse{
					Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
					CredentialSchema: CredentialSchema{
						"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
						"JsonSchemaValidator2018",
					},
					CredentialStatus: verifiable.CredentialStatus{
						ID:              fmt.Sprintf("http://localhost/api/v1/identities/%s/claims/revocation/status/%d", idStr, revNonce),
						Type:            "SparseMerkleTreeProof",
						RevocationNonce: uint64(nonce),
					},
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     float64(19960424),
						"documentType": float64(2),
						"type":         "KYCAgeCredential",
					},
					Id:           fmt.Sprintf("http://localhost/api/v1/claim/%s", claimID),
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
				var responseCredentialSubject, tcCredentialSubject credentialSubject
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				require.Equal(t, response.Id, v.Id)
				require.Equal(t, len(response.Context), len(v.Context))
				require.EqualValues(t, response.Context, v.Context)
				require.EqualValues(t, response.CredentialSchema, v.CredentialSchema)
				require.NoError(t, mapstructure.Decode(response.CredentialSubject, &responseCredentialSubject))
				require.NoError(t, mapstructure.Decode(v.CredentialSubject, &tcCredentialSubject))
				require.EqualValues(t, responseCredentialSubject, tcCredentialSubject)
				require.InDelta(t, response.IssuanceDate.Unix(), v.IssuanceDate.Unix(), 30)
				require.Equal(t, response.Type, v.Type)
				require.NoError(t, mapstructure.Decode(response.CredentialStatus, &responseCredentialStatus))
				credentialStatusTC, ok := v.CredentialStatus.(verifiable.CredentialStatus)
				require.True(t, ok)
				require.EqualValues(t, responseCredentialStatus, credentialStatusTC)
				require.Equal(t, response.Expiration, v.Expiration)
				require.Equal(t, response.Issuer, v.Issuer)

			case GetClaim400JSONResponse:
				var response RevokeClaim404JSONResponse
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
