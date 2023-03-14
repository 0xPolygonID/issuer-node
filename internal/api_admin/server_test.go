package api_admin

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
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func TestServer_CheckStatus(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaService := services.NewSchema(loader.CachedFactory(loader.HTTPFactory, cachex))
	schemaAdminService := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaAdminService, NewConnectionsMock(), NewPublisherMock(), NewPackageManagerMock(), &health.Status{})
	handler := getHandler(context.Background(), server)

	t.Run("should return 200", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/status", nil)
		require.NoError(t, err)

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var response Health200JSONResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	})
}

func TestServer_AuthCallback(t *testing.T) {
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewAdminSchemaMock(), NewConnectionsMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  string
	}
	type testConfig struct {
		name      string
		expected  expected
		sessionID *string
	}

	for _, tc := range []testConfig{
		{
			name: "should get an error empty param sessionID",
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  "Cannot proceed with empty sessionID",
			},
		},
		{
			name:      "should get an error no body",
			sessionID: common.ToPointer("12345"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  "Cannot proceed with empty body",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/authentication/callback"
			if tc.sessionID != nil {
				url += "?sessionID=" + *tc.sessionID
			}

			req, err := http.NewRequest("POST", url, strings.NewReader(``))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusBadRequest:
				var response AuthCallback400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.message, response.Message)
			default:
				t.Fail()
			}
		})
	}
}

func TestServer_AuthQRCode(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	sessionRepository := repositories.NewSessionCached(cachex)

	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, sessionRepository)
	server := NewServer(&cfg, identityService, NewClaimsMock(), NewAdminSchemaMock(), NewConnectionsMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		response AuthQRCodeResponseObject
	}
	type testConfig struct {
		name     string
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "should get a qrCode",
			expected: expected{
				httpCode: http.StatusOK,
				response: AuthQRCode200JSONResponse{
					Body: struct {
						CallbackUrl string        `json:"callbackUrl"`
						Reason      string        `json:"reason"`
						Scope       []interface{} `json:"scope"`
					}{
						CallbackUrl: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       []interface{}{},
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/v1/authentication/qrcode", nil)
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch v := tc.expected.response.(type) {
			case AuthQRCode200JSONResponse:
				var response AuthQRCode200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, v.Typ, response.Typ)
				assert.Equal(t, v.Type, response.Type)
				assert.Equal(t, v.From, response.From)
				assert.Equal(t, v.Body.Scope, response.Body.Scope)
				assert.Equal(t, v.Body.Reason, response.Body.Reason)
				assert.True(t, strings.Contains(response.Body.CallbackUrl, v.Body.CallbackUrl))
			}
		})
	}
}

func TestServer_ImportSchema(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const schemaType = "KYCCountryOfResidenceCredential"
	ctx := context.Background()
	schemaAdminSrv := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaAdminSrv, NewConnectionsMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"

	handler := getHandler(ctx, server)

	type expected struct {
		httpCode int
		errorMsg string
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  *ImportSchemaJSONRequestBody
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:    "Not authorized",
			auth:    authWrong,
			request: nil,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "Empty request",
			auth:    authOk,
			request: nil,
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: empty url",
			},
		},
		{
			name: "Wrong url",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType: "lala",
				Url:        "wrong/url",
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "bad request: parsing url: parse \"wrong/url\": invalid URI for request",
			},
		},
		{
			name: "Valid request",
			auth: authOk,
			request: &ImportSchemaRequest{
				SchemaType: schemaType,
				Url:        url,
			},
			expected: expected{
				httpCode: http.StatusCreated,
				errorMsg: "bad request: parsing url: parse \"wrong/url\": invalid URI for request",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "/v1/schemas", tests.JSONBody(t, tc.request))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response ImportSchema201JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response ImportSchema400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_DeleteConnection(t *testing.T) {
	connectionsRepository := repositories.NewConnections()

	connectionsService := services.NewConnection(connectionsRepository, storage)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaAdminMock(), connectionsService, NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	conn := fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
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
				message:  common.ToPointer("Connection successfully deleted"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/connections/%s", tc.connID.String())
			req, err := http.NewRequest("DELETE", url, nil)
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

func TestServer_CreateCredential(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "mumbai"
	)
	ctx := log.NewContext(context.Background(), log.LevelDebug, log.OutputText, os.Stdout)
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaService := services.NewSchema(loader.CachedFactory(loader.HTTPFactory, cachex))
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewAdminSchemaMock(), connectionsService, NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateCredentialResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		body     CreateCredentialRequest
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
			name: "Happy path",
			auth: authOk,
			body: CreateCredentialRequest{
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
				response: CreateCredential201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Wrong credential url",
			auth: authOk,
			body: CreateCredentialRequest{
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
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "malformed url"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Unreachable well formed credential url",
			auth: authOk,
			body: CreateCredentialRequest{
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
				response: CreateCredential422JSONResponse{N422JSONResponse{Message: "cannot load schema"}},
				httpCode: http.StatusUnprocessableEntity,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/credentials"

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response CreateCredentialResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
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

func TestServer_GetCredential(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "mumbai"
	)
	ctx := log.NewContext(context.Background(), log.LevelDebug, log.OutputText, os.Stdout)
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaService := services.NewSchema(loader.CachedFactory(loader.HTTPFactory, cachex))
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, schemaService, connectionsService, NewPublisherMock(), NewPackageManagerMock(), nil)

	fixture := tests.NewFixture(storage)
	claim := fixture.NewClaim(t, did.String())
	fixture.CreateClaim(t, claim)

	handler := getHandler(ctx, server)

	type credentialKYCSubject struct {
		Id           string `json:"id"`
		Birthday     uint64 `json:"birthday"`
		DocumentType uint64 `json:"documentType"`
		Type         string `json:"type"`
	}

	type expected struct {
		message  *string
		response GetCredential200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  GetCredentialRequestObject
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			request: GetCredentialRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should return an error, claim not found",
			auth: authOk,
			request: GetCredentialRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				message:  common.ToPointer("The given credential id does not exist"),
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "happy path",
			auth: authOk,
			request: GetCredentialRequestObject{
				Id: claim.ID,
			},
			expected: expected{
				response: GetCredential200JSONResponse{
					Attributes: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
						"type":         "KYCAgeCredential",
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         claim.ID,
					ProofTypes: []string{},
					RevNonce:   uint64(claim.RevNonce),
					Revoked:    claim.Revoked,
					SchemaHash: claim.SchemaHash,
					SchemaType: claim.SchemaType,
				},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/%s", tc.request.Id.String())

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetCredentialResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.Id, response.Id)
				assert.Equal(t, tc.expected.response.SchemaType, response.SchemaType)
				assert.Equal(t, tc.expected.response.SchemaHash, response.SchemaHash)
				assert.Equal(t, tc.expected.response.Revoked, response.Revoked)
				assert.Equal(t, tc.expected.response.RevNonce, response.RevNonce)
				assert.InDelta(t, tc.expected.response.CreatedAt.Unix(), response.CreatedAt.Unix(), 10)
				if response.ExpiresAt != nil && tc.expected.response.ExpiresAt != nil {
					assert.InDelta(t, tc.expected.response.ExpiresAt.Unix(), response.ExpiresAt.Unix(), 10)
				}
				assert.Equal(t, tc.expected.response.Expired, response.Expired)
				var respAttributes, tcCredentialSubject credentialKYCSubject
				assert.NoError(t, mapstructure.Decode(tc.expected.response.Attributes, &tcCredentialSubject))
				assert.NoError(t, mapstructure.Decode(response.Attributes, &respAttributes))
				assert.EqualValues(t, respAttributes, tcCredentialSubject)
				assert.EqualValues(t, tc.expected.response.ProofTypes, response.ProofTypes)
			case http.StatusBadRequest:
				var response GetCredential400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}
