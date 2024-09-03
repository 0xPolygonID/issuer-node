package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestServer_RevokeClaim(t *testing.T) {
	server := newTestServer(t, nil)

	idStr := "did:polygonid:polygon:mumbai:2qM77fA6NGGWL9QEeb1dv2VA6wz5svcohgv61LZ7wB"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture := repositories.NewFixture(storage)
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

	query := repositories.ExecQueryParams{
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
		response RevokeCredentialResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "No auth header",
			auth:  authWrong,
			did:   idStr,
			nonce: nonce,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should revoke the credentials",
			auth:  authOk,
			did:   idStr,
			nonce: nonce,
			expected: expected{
				httpCode: 202,
				response: RevokeCredential202JSONResponse{
					Message: "credential revocation request sent",
				},
			},
		},
		{
			name:  "should get an error wrong nonce",
			auth:  authOk,
			did:   idStr,
			nonce: int64(1231323),
			expected: expected{
				httpCode: 404,
				response: RevokeCredential404JSONResponse{N404JSONResponse{
					Message: "the credential does not exist",
				}},
			},
		},
		{
			name:  "should get an error",
			auth:  authOk,
			did:   "did:polygonid:polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			nonce: nonce,
			expected: expected{
				httpCode: 500,
				response: RevokeCredential500JSONResponse{N500JSONResponse{
					Message: "error getting merkle trees: not found",
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/revoke/%d", tc.did, tc.nonce)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case RevokeCredential202JSONResponse:
				var response RevokeCredential202JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case RevokeCredential404JSONResponse:
				var response RevokeCredential404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case RevokeCredential500JSONResponse:
				var response RevokeCredential500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			}
		})
	}
}

func TestServer_CreateCredential(t *testing.T) {
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

	claimID, err := uuid.NewUUID()
	require.NoError(t, err)

	type expected struct {
		response                    CreateCredentialResponseObject
		httpCode                    int
		createCredentialEventsCount int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		body     CreateCredentialRequest
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			did:  did,
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with two proofs",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateCredentialRequestProofs{
					"BJJSignature2021",
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with bjjSignature proof",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateCredentialRequestProofs{
					"BJJSignature2021",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with Iden3SparseMerkleTreeProof proof",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateCredentialRequestProofs{
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 0,
			},
		},
		{
			name: "Happy path with ipfs schema",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL",
				Type:             "testNewType",
				CredentialSubject: map[string]any{
					"id":             "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"testNewTypeInt": 1234,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with refresh service",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL",
				Type:             "testNewType",
				CredentialSubject: map[string]any{
					"id":             "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"testNewTypeInt": 1234,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				RefreshService: &RefreshService{
					Id:   "http://localhost:8080",
					Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
				},
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with credentials id",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				ClaimID:          common.ToPointer(claimID),
				CredentialSchema: "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qFDkNkWePjd6URt6kGQX14a7wVKhBZt8bpy7HZJZi",
					"birthday":     19960425,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs: &[]CreateCredentialRequestProofs{
					"BJJSignature2021",
					"Iden3SparseMerkleTreeProof",
				},
			},
			expected: expected{
				response: CreateCredential201JSONResponse{
					Id: claimID.String(),
				},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Wrong credential url",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "wrong url",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "malformed url"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Unreachable well formed credential url",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "http://www.wrong.url/cannot/get/the/credential",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
			},
			expected: expected{
				response: CreateCredential422JSONResponse{N422JSONResponse{Message: "cannot load schema"}},
				httpCode: http.StatusUnprocessableEntity,
			},
		},
		{
			name: "Wrong proof type",
			auth: authOk,
			did:  did,
			body: CreateCredentialRequest{
				CredentialSchema: "http://www.wrong.url/cannot/get/the/credential",
				Type:             "KYCAgeCredential",
				CredentialSubject: map[string]any{
					"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"birthday":     19960424,
					"documentType": 2,
				},
				Expiration: common.ToPointer(time.Now().Unix()),
				Proofs:     &[]CreateCredentialRequestProofs{"wrong proof"},
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "unsupported proof type: wrong proof"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server.Infra.pubSub.Clear(event.CreateCredentialEvent)
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials", tc.did)

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			assert.Equal(t, tc.expected.createCredentialEventsCount, len(server.Infra.pubSub.AllPublishedEvents(event.CreateCredentialEvent)))

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateCredentialResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
				if tc.body.ClaimID != nil {
					assert.Equal(t, tc.body.ClaimID.String(), response.Id)
				}
			case http.StatusBadRequest:
				var response CreateCredential400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			case http.StatusUnprocessableEntity:
				var response CreateCredential422JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_DeleteCredential(t *testing.T) {
	server := newTestServer(t, nil)
	ctx := context.Background()
	handler := getHandler(ctx, server)
	identity, err := server.Services.identity.Create(ctx, "http://polygon-test", &ports.DIDCreationOptions{Method: core.DIDMethodIden3, Blockchain: core.Polygon, Network: core.Amoy, KeyType: kms.KeyTypeBabyJubJub})
	require.NoError(t, err)
	fixture := repositories.NewFixture(storage)
	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

	type expected struct {
		httpCode int
		message  *string
	}

	type testConfig struct {
		name         string
		credentialID uuid.UUID
		auth         func() (string, string)
		expected     expected
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
			name:         "should get an error, not existing claim",
			credentialID: uuid.New(),
			auth:         authOk,
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("The given credential does not exist"),
			},
		},
		{
			name:         "should delete the credential",
			credentialID: claim.ID,
			auth:         authOk,
			expected: expected{
				httpCode: http.StatusOK,
				message:  common.ToPointer("Credential successfully deleted"),
			},
		},
		{
			name:         "should get an error, a credential cannot be deleted twice",
			credentialID: claim.ID,
			auth:         authOk,
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("The given credential does not exist"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/%s", *claim.Identifier, tc.credentialID.String())
			req, err := http.NewRequest("DELETE", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response DeleteCredential400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			case http.StatusOK:
				var response DeleteCredential200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_GetCredentialQrCode(t *testing.T) {
	idStr := "did:polygonid:polygon:mumbai:2qPrv5Yx8s1qAmEnPym68LfT7gTbASGampiGU7TseL"
	idNoClaims := "did:polygonid:polygon:mumbai:2qGjTUuxZKqKS4Q8UmxHUPw55g15QgEVGnj6Wkq8Vk"
	identity := &domain.Identity{
		Identifier: idStr,
	}

	fixture := repositories.NewFixture(storage)
	fixture.CreateIdentity(t, identity)
	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

	server := newTestServer(t, nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		response GetCredentialQrCodeResponseObject
		httpCode int
		qrType   string
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		claim    uuid.UUID
		qrType   *string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:  "No auth",
			auth:  authWrong,
			did:   idStr,
			claim: claim.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should get an error non existing claimID",
			auth:  authOk,
			did:   idStr,
			claim: uuid.New(),
			expected: expected{
				response: GetCredentialQrCode404JSONResponse{N404JSONResponse{
					Message: "Credential not found",
				}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name:  "should get an error the given did has no entry for claimID",
			auth:  authOk,
			did:   idNoClaims,
			claim: claim.ID,
			expected: expected{
				response: GetCredentialQrCode404JSONResponse{N404JSONResponse{
					Message: "Credential not found",
				}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name:  "should get an error wrong did invalid format",
			auth:  authOk,
			did:   ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			claim: claim.ID,
			expected: expected{
				response: GetCredentialQrCode400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Happy path, no type",
			auth:   authOk,
			did:    idStr,
			claim:  claim.ID,
			qrType: nil,
			expected: expected{
				response: GetCredentialQrCode200JSONResponse{},
				httpCode: http.StatusOK,
				qrType:   "universalLink",
			},
		},
		{
			name:   "Happy path, type universalLink",
			auth:   authOk,
			did:    idStr,
			claim:  claim.ID,
			qrType: common.ToPointer("universalLink"),
			expected: expected{
				response: GetCredentialQrCode200JSONResponse{},
				httpCode: http.StatusOK,
				qrType:   "universalLink",
			},
		},
		{
			name:   "Happy path, type deeplink",
			auth:   authOk,
			did:    idStr,
			claim:  claim.ID,
			qrType: common.ToPointer("deepLink"),
			expected: expected{
				response: GetCredentialQrCode200JSONResponse{},
				httpCode: http.StatusOK,
				qrType:   "deepLink",
			},
		},
		{
			name:   "Happy path, type raw",
			auth:   authOk,
			did:    idStr,
			claim:  claim.ID,
			qrType: common.ToPointer("raw"),
			expected: expected{
				response: GetCredentialQrCode200JSONResponse{},
				httpCode: http.StatusOK,
				qrType:   "raw",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/%s/qrcode", tc.did, tc.claim)
			if tc.qrType != nil {
				url = fmt.Sprintf("%s?type=%s", url, *tc.qrType)
			}
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetCredentialQrCode200JSONResponse:
				var response GetCredentialQrCode200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				switch tc.expected.qrType {
				case "raw":
					var rawResponse QrCodeLinkWithSchemaTypeShortResponse
					assert.NoError(t, json.Unmarshal([]byte(response.UniversalLink), &rawResponse))
				case "universalLink":
					assert.True(t, strings.HasPrefix(response.UniversalLink, "https://testing.env#request_uri=https://testing.env/v2/qr-store?id="))
				case "deepLink":
					assert.True(t, strings.HasPrefix(response.UniversalLink, "iden3comm://?request_uri=https://testing.env/v2/qr-store?id="))
				}
			case GetCredentialQrCode400JSONResponse:
				var response GetCredentialQrCode400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredentialQrCode404JSONResponse:
				var response GetCredentialQrCode400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredentialQrCode500JSONResponse:
				var response GetCredentialQrCode400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			}
		})
	}
}

func TestServer_GetCredential(t *testing.T) {
	server := newTestServer(t, nil)
	idStr := "did:polygonid:polygon:mumbai:2qLduMv2z7hnuhzkcTWesCUuJKpRVDEThztM4tsJUj"
	idStrWithoutClaims := "did:polygonid:polygon:mumbai:2qGjTUuxZKqKS4Q8UmxHUPw55g15QgEVGnj6Wkq8Vk"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture := repositories.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	claim := fixture.NewClaim(t, identity.Identifier)
	fixture.CreateClaim(t, claim)

	query := repositories.ExecQueryParams{
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
		response GetCredentialResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		did      string
		claimID  uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			did:  idStr,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "should get an error non existing claimID",
			auth:    authOk,
			did:     idStr,
			claimID: uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				response: GetCredential404JSONResponse{N404JSONResponse{
					Message: "credential not found",
				}},
			},
		},
		{
			name:    "should get an error the given did has no entry for claimID",
			auth:    authOk,
			did:     idStrWithoutClaims,
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusNotFound,
				response: GetCredential404JSONResponse{N404JSONResponse{
					Message: "credential not found",
				}},
			},
		},
		{
			name:    "should get an error wrong did invalid format",
			auth:    authOk,
			did:     ":polygon:mumbai:2qPUUYXa98tQWZKSaRidf2QTDyZicFFxkTWNWjk2HJ",
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusBadRequest,
				response: GetCredential400JSONResponse{N400JSONResponse{
					Message: "invalid did",
				}},
			},
		},
		{
			name:    "should get the credentials",
			auth:    authOk,
			did:     idStr,
			claimID: claim.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetCredential200JSONResponse{
					Context: []string{"https://www.w3.org/2018/credentials/v1", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/iden3credential-v2.json-ld", "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"},
					CredentialSchema: CredentialSchema{
						"https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json",
						"JsonSchemaValidator2018",
					},
					CredentialStatus: verifiable.CredentialStatus{
						ID:              fmt.Sprintf("http://localhost/v2/%s/credentials/revocation/status/%d", idStr, claim.RevNonce),
						Type:            "SparseMerkleTreeProof",
						RevocationNonce: uint64(claim.RevNonce),
					},
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     float64(19960424),
						"documentType": float64(2),
						"type":         "KYCAgeCredential",
					},
					Id:           fmt.Sprintf("http://localhost/api/v2/credentials/%s", claim.ID),
					IssuanceDate: common.ToPointer(TimeUTC(time.Now())),
					Issuer:       idStr,
					Type:         []string{"VerifiableCredential", "KYCAgeCredential"},
					RefreshService: &RefreshService{
						Id:   "https://refresh-service.xyz",
						Type: RefreshServiceType(verifiable.Iden3RefreshService2023),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/%s", tc.did, tc.claimID.String())
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case GetCredential200JSONResponse:
				var response GetCredentialResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				validateCredential(t, response, GetCredentialResponse(v))

			case GetCredential400JSONResponse:
				var response GetCredential404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredential404JSONResponse:
				var response GetCredential404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			case GetCredential500JSONResponse:
				var response GetCredential500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Message, response.Message)
			}
		})
	}
}

func TestServer_GetCredentials(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()

	server := newTestServer(t, nil)
	identityMultipleClaims, err := server.identityService.Create(ctx, "https://localhost.com", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(identityMultipleClaims.Identifier)
	require.NoError(t, err)

	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schemaURL := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}

	claimsService := server.claimService

	// Never expires
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true},
		nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	future := time.Now().Add(1000 * time.Hour)
	past := time.Now().Add(-1000 * time.Hour)
	// Expires in future
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	// Expired
	expiredClaim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &past, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	// non expired, but revoked
	revoked, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition,
		ports.ClaimRequestProofs{BJJSignatureProof2021: false, Iden3SparseMerkleTreeProof: true},
		nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	id, err := w3c.ParseDID(*revoked.Identifier)
	require.NoError(t, err)
	require.NoError(t, claimsService.Revoke(ctx, *id, uint64(revoked.RevNonce), "because I can"))

	iReq := ports.NewImportSchemaRequest(schemaURL, typeC, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	_, err = server.schemaService.ImportSchema(ctx, *did, iReq)
	require.NoError(t, err)

	handler := getHandler(context.Background(), server)

	type expected struct {
		credentialsCount int
		page             uint
		maxResults       uint
		total            uint
		httpCode         int
		errorMsg         string
	}

	type testConfig struct {
		name       string
		auth       func() (string, string)
		did        *string
		query      *string
		sort       *string
		status     *string
		page       *int
		maxResults *int
		expected   expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			page: common.ToPointer(1),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "Wrong status",
			auth:   authOk,
			page:   common.ToPointer(1),
			status: common.ToPointer("wrong"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "wrong type value. Allowed values: [all, revoked, expired]",
			},
		},
		{
			name: "wrong did",
			auth: authOk,
			page: common.ToPointer(1),
			did:  common.ToPointer("wrongdid:"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "cannot parse did parameter: wrong format",
			},
		},
		{
			name: "pagination. Page is < 1 not allowed",
			auth: authOk,
			page: common.ToPointer(0),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "page param must be higher than 0",
			},
		},
		{
			name:       "pagination. max_results < 1 return default max results",
			auth:       authOk,
			page:       common.ToPointer(1),
			maxResults: common.ToPointer(0),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name: "Default max results",
			auth: authOk,
			page: common.ToPointer(1),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:   "Status 'all' explicit",
			auth:   authOk,
			page:   common.ToPointer(1),
			status: common.ToPointer("all"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:       "GetCredentialsPaginated all explicit, page 1 with 2 results",
			auth:       authOk,
			status:     common.ToPointer("all"),
			page:       common.ToPointer(1),
			maxResults: common.ToPointer(2),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       2,
				page:             1,
				credentialsCount: 2,
			},
		},
		{
			name:       "GetCredentialsPaginated all explicit, page 2 with 2 results",
			auth:       authOk,
			status:     common.ToPointer("all"),
			page:       common.ToPointer(2),
			maxResults: common.ToPointer(2),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       2,
				page:             2,
				credentialsCount: 2,
			},
		},
		{
			name:       "GetCredentialsPaginated all explicit, page 3 with 2 results. No results",
			auth:       authOk,
			status:     common.ToPointer("all"),
			page:       common.ToPointer(3),
			maxResults: common.ToPointer(2),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       2,
				page:             3,
				credentialsCount: 0,
			},
		},
		{
			name:   "GetCredentialsPaginated all from existing did",
			auth:   authOk,
			status: common.ToPointer("all"),
			page:   common.ToPointer(1),
			did:    &expiredClaim.OtherIdentifier,
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:   "GetCredentialsPaginated all from non existing did. Expecting empty list",
			auth:   authOk,
			status: common.ToPointer("all"),
			page:   common.ToPointer(1),
			did:    common.ToPointer("did:iden3:tJU7z1dbKyKYLiaopZ5tN6Zjsspq7QhYayiR31RFa"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            0,
				maxResults:       50,
				page:             1,
				credentialsCount: 0,
			},
		},
		{
			name:   "Revoked",
			auth:   authOk,
			status: common.ToPointer("revoked"),
			page:   common.ToPointer(1),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            1,
				maxResults:       50,
				page:             1,
				credentialsCount: 1,
			},
		},
		{
			name:   "REVOKED",
			auth:   authOk,
			status: common.ToPointer("REVOKED"),
			page:   common.ToPointer(1),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            1,
				maxResults:       50,
				page:             1,
				credentialsCount: 1,
			},
		},
		{
			name:   "Expired",
			auth:   authOk,
			status: common.ToPointer("expired"),
			page:   common.ToPointer(1),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            1,
				maxResults:       50,
				page:             1,
				credentialsCount: 1,
			},
		},
		{
			name:  "Search by did and other words in query params:",
			auth:  authOk,
			page:  common.ToPointer(1),
			query: common.ToPointer("some words and " + revoked.OtherIdentifier),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "Search by partial did and other words in query params:",
			auth:  authOk,
			page:  common.ToPointer(1),
			query: common.ToPointer("some words and " + revoked.OtherIdentifier[9:14]),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "Search by did in query params:",
			auth:  authOk,
			page:  common.ToPointer(1),
			query: &revoked.OtherIdentifier,
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "Search by attributes in query params",
			auth:  authOk,
			query: common.ToPointer("birthday"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "Search by attributes in query params, partial word",
			auth:  authOk,
			page:  common.ToPointer(1),
			query: common.ToPointer("rthd"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "Search by partial did in query params:",
			auth:  authOk,
			query: common.ToPointer(revoked.OtherIdentifier[9:14]),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "FTS is doing and OR when no did passed:",
			auth:  authOk,
			query: common.ToPointer("birthday schema attribute not the rest of words this sentence"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:  "FTS is doing and AND when did passed:",
			auth:  authOk,
			did:   &expiredClaim.OtherIdentifier,
			query: common.ToPointer("not existing words"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            0,
				maxResults:       50,
				page:             1,
				credentialsCount: 0,
			},
		},
		{
			name: "Wrong order by",
			auth: authOk,
			sort: common.ToPointer("wrongField"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "wrong sort by value",
			},
		},
		{
			name: "Order by one field",
			auth: authOk,
			sort: common.ToPointer("createdAt"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name: "Order by 2 fields",
			auth: authOk,
			sort: common.ToPointer("-schemaType, createdAt"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name: "Order by all fields",
			auth: authOk,
			sort: common.ToPointer("-schemaType, createdAt, -expiresAt, revoked"),
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name: "Order by 2 repeated fields",
			auth: authOk,
			sort: common.ToPointer("createdAt, createdAt"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "repeated sort by value field",
			},
		},
		{
			name: "Order by 2 repeated contradictory fields ",
			auth: authOk,
			sort: common.ToPointer("createdAt, -createdAt"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "repeated sort by value field",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := url.URL{Path: fmt.Sprintf("/v2/identities/%s/credentials/search", identityMultipleClaims.Identifier)}
			queryParams := make([]string, 0)
			if tc.query != nil {
				queryParams = append(queryParams, "query="+*tc.query)
			}
			if tc.sort != nil {
				queryParams = append(queryParams, "sort="+*tc.sort)
			}
			if tc.status != nil {
				queryParams = append(queryParams, "status="+*tc.status)
			}
			if tc.did != nil {
				queryParams = append(queryParams, "credentialSubject="+*tc.did)
			}
			if tc.page != nil {
				queryParams = append(queryParams, "page="+strconv.Itoa(*tc.page))
			}
			if tc.maxResults != nil {
				queryParams = append(queryParams, "max_results="+strconv.Itoa(*tc.maxResults))
			}
			endpoint.RawQuery = strings.Join(queryParams, "&")
			req, err := http.NewRequest("GET", endpoint.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetCredentials200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.total, response.Meta.Total)
				assert.Equal(t, tc.expected.credentialsCount, len(response.Items))
				assert.Equal(t, tc.expected.maxResults, response.Meta.MaxResults)
				assert.Equal(t, tc.expected.page, response.Meta.Page)

			case http.StatusBadRequest:
				var response GetCredentials400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)

			case http.StatusInternalServerError:
				var response GetCredentials400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_GetRevocationStatus(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()

	server := newTestServer(t, nil)
	identity, err := server.Services.identity.Create(context.Background(), "http://localhost:3001", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)
	handler := getHandler(context.Background(), server)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, _ := w3c.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "value"
	claimRequestProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: true,
	}
	credential, err := server.Services.credentials.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, claimRequestProofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	assert.NoError(t, err)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "should get revocation status",
			auth:  authOk,
			nonce: int64(credential.RevNonce),
			expected: expected{
				httpCode: http.StatusOK,
			},
		},

		{
			name:  "should get revocation status wrong nonce",
			auth:  authOk,
			nonce: 123456,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/%s/claims/revocation/status/%d", identity.Identifier, tc.nonce)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			if tc.expected.httpCode == http.StatusOK {
				var response GetRevocationStatus200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.Issuer.ClaimsTreeRoot)
				assert.NotNil(t, response.Issuer.State)
				assert.NotNil(t, response.Mtp.Existence)
				assert.NotNil(t, response.Mtp.Siblings)
			}
		})
	}
}

func TestServer_GetRevocationStatusV2(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()

	server := newTestServer(t, nil)
	identity, err := server.Services.identity.Create(context.Background(), "http://localhost:3001", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)
	handler := getHandler(context.Background(), server)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, _ := w3c.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "value"
	claimRequestProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: true,
	}
	credential, err := server.Services.credentials.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, claimRequestProofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	assert.NoError(t, err)

	type expected struct {
		httpCode int
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		nonce    int64
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:  "should get revocation status",
			auth:  authOk,
			nonce: int64(credential.RevNonce),
			expected: expected{
				httpCode: http.StatusOK,
			},
		},

		{
			name:  "should get revocation status wrong nonce",
			auth:  authOk,
			nonce: 123456,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v2/identities/%s/credentials/revocation/status/%d", identity.Identifier, tc.nonce)
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			if tc.expected.httpCode == http.StatusOK {
				var response GetRevocationStatusV2200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.Issuer.ClaimsTreeRoot)
				assert.NotNil(t, response.Issuer.State)
				assert.NotNil(t, response.Mtp.Existence)
				assert.NotNil(t, response.Mtp.Siblings)
			}
		})
	}
}

func validateCredential(t *testing.T, resp, tc GetCredentialResponse) {
	t.Helper()
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
	assert.InDelta(t, time.Time(*resp.IssuanceDate).UnixMilli(), time.Time(*tc.IssuanceDate).UnixMilli(), 1000)
	assert.Equal(t, resp.Type, tc.Type)
	assert.Equal(t, resp.ExpirationDate, tc.ExpirationDate)
	assert.Equal(t, resp.Issuer, tc.Issuer)
	assert.Equal(t, resp.RefreshService, tc.RefreshService)
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
