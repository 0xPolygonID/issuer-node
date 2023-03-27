package api_admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	schemaAdminService := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaAdminService, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), &health.Status{})
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
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewAdminSchemaMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  string
	}
	type testConfig struct {
		name      string
		expected  expected
		sessionID *uuid.UUID
	}

	for _, tc := range []testConfig{
		{
			name:      "should get an error no body",
			sessionID: common.ToPointer(uuid.New()),
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
				url += "?sessionID=" + tc.sessionID.String()
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
	server := NewServer(&cfg, identityService, NewClaimsMock(), NewAdminSchemaMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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

func TestServer_GetSchema(t *testing.T) {
	ctx := context.Background()
	schemaAdminSrv := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaAdminSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	fixture := tests.NewFixture(storage)

	s := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		URL:        "https://domain.org/this/is/an/url",
		Type:       "schemaType",
		Attributes: domain.SchemaAttrsFromString("attr1, attr2, attr3"),
		CreatedAt:  time.Now(),
	}
	s.Hash = utils.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
	fixture.CreateSchema(t, ctx, s)
	sHash, _ := s.Hash.MarshalText()

	handler := getHandler(ctx, server)
	type expected struct {
		httpCode int
		errorMsg string
		schema   *Schema
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		id       string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			id:   uuid.NewString(),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Invalid uuid",
			auth: authOk,
			id:   "someInvalidDID",
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "Invalid format for parameter id: error unmarshalling 'someInvalidDID' text as *uuid.UUID: invalid UUID length: 14",
			},
		},
		{
			name: "Non existing uuid",
			auth: authOk,
			id:   uuid.NewString(),
			expected: expected{
				httpCode: http.StatusNotFound,
				errorMsg: "schema not found",
			},
		},
		{
			name: "Happy path. Existing schema",
			auth: authOk,
			id:   s.ID.String(),
			expected: expected{
				httpCode: http.StatusOK,
				schema: &Schema{
					BigInt:    s.Hash.BigInt().String(),
					CreatedAt: s.CreatedAt,
					Hash:      string(sHash),
					Id:        s.ID.String(),
					Type:      s.Type,
					Url:       s.URL,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", fmt.Sprintf("/v1/schemas/%s", tc.id), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetSchema200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.schema.Id, response.Id)
				assert.Equal(t, tc.expected.schema.BigInt, response.BigInt)
				assert.Equal(t, tc.expected.schema.Type, response.Type)
				assert.Equal(t, tc.expected.schema.Url, response.Url)
				assert.Equal(t, tc.expected.schema.Hash, response.Hash)
				assert.InDelta(t, tc.expected.schema.CreatedAt.UnixMilli(), response.CreatedAt.UnixMilli(), 10)
			case http.StatusNotFound:
				var response GetSchema404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			case http.StatusBadRequest:
				var response GetSchema400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

// Refer to the schema repository tests for more deep test related to Postgres Full Text Search
func TestServer_GetSchemas(t *testing.T) {
	ctx := context.Background()
	// Need an isolated DB here to avoid other tests side effects
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}
	storage, teardown, err := tests.NewTestStorage(&config.Configuration{Database: config.Database{URL: conn}})
	require.NoError(t, err)
	defer teardown()

	schemaAdminSrv := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaAdminSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	fixture := tests.NewFixture(storage)

	for i := 0; i < 20; i++ {
		s := &domain.Schema{
			ID:         uuid.New(),
			IssuerDID:  *issuerDID,
			URL:        fmt.Sprintf("https://domain.org/this/is/an/url/%d", i),
			Type:       fmt.Sprintf("schemaType-%d", i),
			Attributes: domain.SchemaAttrsFromString("attr1, attr2, attr3"),
			CreatedAt:  time.Now(),
		}
		s.Hash = utils.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
		fixture.CreateSchema(t, ctx, s)
	}
	s := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		URL:        "https://domain.org/this/is/an/url/ubiprogram",
		Type:       "UbiProgram",
		Attributes: domain.SchemaAttrsFromString("attr1, attr2, attr3"),
		CreatedAt:  time.Now(),
	}
	s.Hash = utils.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
	fixture.CreateSchema(t, ctx, s)

	handler := getHandler(ctx, server)
	type expected struct {
		httpCode int
		count    int
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		query    *string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "Happy path. All schemas, no query",
			auth:  authOk,
			query: nil,
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
			},
		},
		{
			name:  "Happy path. All schemas, query=''",
			auth:  authOk,
			query: common.ToPointer(""),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
			},
		},
		{
			name:  "Happy path. Search for schema type. All",
			auth:  authOk,
			query: common.ToPointer("schemaType"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    20,
			},
		},
		{
			name:  "Happy path. Search for one schema but many attr. Return all",
			auth:  authOk,
			query: common.ToPointer("schemaType-11 attr1"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    21,
			},
		},
		{
			name:  "Exact search, return 1",
			auth:  authOk,
			query: common.ToPointer("UbiProgram"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := "/v1/schemas"
			if tc.query != nil {
				endpoint = endpoint + "?query=" + url.QueryEscape(*tc.query)
			}
			req, err := http.NewRequest("GET", endpoint, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetSchemas200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, len(response), tc.expected.count)
			}
		})
	}
}

func TestServer_ImportSchema(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const schemaType = "KYCCountryOfResidenceCredential"
	ctx := context.Background()
	schemaAdminSrv := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaAdminSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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

func TestServer_DeleteConnectionCredentials(t *testing.T) {
	connectionsRepository := repositories.NewConnections()

	connectionsService := services.NewConnection(connectionsRepository, storage)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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
			url := fmt.Sprintf("/v1/connections/%s/credentials", tc.connID.String())
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	issuerDID, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	server.cfg.APIUI.IssuerDID = *issuerDID
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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
		RevNonce:        0,
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
			url := fmt.Sprintf("/v1/connections/%s/credentials/revoke", tc.connID.String())
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewAdminSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

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
				Expiration:     common.ToPointer(int64(12345)),
				SignatureProof: common.ToPointer(true),
			},
			expected: expected{
				response: CreateCredential201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Wrong request - no proof provided",
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
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "you must to provide at least one proof type"}},
				httpCode: http.StatusBadRequest,
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
				Expiration:     common.ToPointer(int64(12345)),
				SignatureProof: common.ToPointer(true),
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
				Expiration:     common.ToPointer(int64(12345)),
				SignatureProof: common.ToPointer(true),
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
				var response UUIDResponse
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

func TestServer_DeleteCredential(t *testing.T) {
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)

	cred := fixture.NewClaim(t, issuerDID.String())
	fCred := fixture.CreateClaim(t, cred)

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
			credentialID: fCred,
			auth:         authOk,
			expected: expected{
				httpCode: http.StatusOK,
				message:  common.ToPointer("Credential successfully deleted"),
			},
		},
		{
			name:         "should get an error, a credential can not be deleted twice",
			credentialID: fCred,
			auth:         authOk,
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("The given credential does not exist"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/%s", tc.credentialID.String())
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	createdClaim1, err := claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true)))
	require.NoError(t, err)

	createdClaim2, err := claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(false)))
	require.NoError(t, err)

	createdClaim3, err := claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(false), common.ToPointer(true)))
	require.NoError(t, err)
	handler := getHandler(ctx, server)

	type expected struct {
		message  *string
		response Credential
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
			name: "happy path with two proof",
			auth: authOk,
			request: GetCredentialRequestObject{
				Id: createdClaim1.ID,
			},
			expected: expected{
				response: Credential{
					Attributes: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
						"type":         "KYCAgeCredential",
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim1.ID,
					ProofTypes: []string{"BJJSignature2021", "MTP"},
					RevNonce:   uint64(createdClaim1.RevNonce),
					Revoked:    createdClaim1.Revoked,
					SchemaHash: createdClaim1.SchemaHash,
					SchemaType: createdClaim1.SchemaType,
				},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "happy path with signature proof",
			auth: authOk,
			request: GetCredentialRequestObject{
				Id: createdClaim2.ID,
			},
			expected: expected{
				response: Credential{
					Attributes: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
						"type":         "KYCAgeCredential",
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim2.ID,
					ProofTypes: []string{"BJJSignature2021"},
					RevNonce:   uint64(createdClaim2.RevNonce),
					Revoked:    createdClaim2.Revoked,
					SchemaHash: createdClaim2.SchemaHash,
					SchemaType: createdClaim2.SchemaType,
				},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "happy path with MTP proof",
			auth: authOk,
			request: GetCredentialRequestObject{
				Id: createdClaim3.ID,
			},
			expected: expected{
				response: Credential{
					Attributes: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
						"type":         "KYCAgeCredential",
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim3.ID,
					ProofTypes: []string{"MTP"},
					RevNonce:   uint64(createdClaim3.RevNonce),
					Revoked:    createdClaim3.Revoked,
					SchemaHash: createdClaim3.SchemaHash,
					SchemaType: createdClaim3.SchemaType,
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
				var response Credential
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				validateCredential(t, tc.expected.response, response)
			case http.StatusBadRequest:
				var response GetCredential400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_GetCredentials(t *testing.T) {
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	day0 := time.Time{}.Unix()
	future := time.Now().Add(1000 * time.Hour).Unix()
	_, err = claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, &day0, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true)))
	require.NoError(t, err)

	_, err = claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(false)))
	require.NoError(t, err)

	revoked, err := claimsService.CreateClaim(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(false), common.ToPointer(true)))
	require.NoError(t, err)

	id, err := core.ParseDID(*revoked.Identifier)
	require.NoError(t, err)
	require.NoError(t, claimsService.Revoke(ctx, *id, uint64(revoked.RevNonce), "because I can"))

	handler := getHandler(ctx, server)

	type expected struct {
		count    int
		httpCode int
		errorMsg string
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		query    *string
		rType    *string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "Wrong type",
			auth:  authOk,
			rType: common.ToPointer("wrong"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "Wrong type value. Allowed values: [all, revoked, expired]",
			},
		},
		{
			name: "Get all implicit",
			auth: authOk,
			expected: expected{
				httpCode: http.StatusOK,
				count:    3,
			},
		},
		{
			name:  "Get all explicit",
			auth:  authOk,
			rType: common.ToPointer("all"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    3,
			},
		},
		{
			name:  "Revoked",
			auth:  authOk,
			rType: common.ToPointer("revoked"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:  "REVOKED",
			auth:  authOk,
			rType: common.ToPointer("REVOKED"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:  "Expired",
			auth:  authOk,
			rType: common.ToPointer("expired"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:  "Search by did:",
			auth:  authOk,
			query: common.ToPointer("some words and " + revoked.OtherIdentifier),
			expected: expected{
				httpCode: http.StatusOK,
				count:    3,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := url.URL{Path: "/v1/credentials"}
			if tc.query != nil {
				endpoint.RawQuery = endpoint.RawQuery + "query=" + *tc.query
			}
			if tc.rType != nil {
				endpoint.RawQuery = endpoint.RawQuery + "type=" + *tc.rType
			}
			req, err := http.NewRequest("GET", endpoint.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetCredentials200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Len(t, response, tc.expected.count)
			case http.StatusBadRequest:
				var response GetCredentials400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_GetConnection(t *testing.T) {
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	fixture := tests.NewFixture(storage)
	claim := fixture.NewClaim(t, did.String())
	fixture.CreateClaim(t, claim)

	usrDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	usrDID2, err := core.ParseDID("did:polygonid:polygon:mumbai:2qFBp1sRF1bFbTybVHHZQRgSWE2nKrdWeAxyZ67PdG")
	require.NoError(t, err)

	connID := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *did,
		UserDID:    *usrDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	connID2 := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *did,
		UserDID:    *usrDID2,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	handler := getHandler(ctx, server)

	type expected struct {
		message  *string
		response GetConnection200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  GetConnectionRequestObject
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			request: GetConnectionRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "should return an error, connection not found",
			auth: authOk,
			request: GetConnectionRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				message:  common.ToPointer("The given connection does not exist"),
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "happy path 1 claim",
			auth: authOk,
			request: GetConnectionRequestObject{
				Id: connID,
			},
			expected: expected{
				response: GetConnection200JSONResponse{
					Connection: Connection{
						CreatedAt: time.Now(),
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
					},
					Credentials: common.ToPointer([]Credential{
						{
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
					}),
				},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "happy path 0 claims",
			auth: authOk,
			request: GetConnectionRequestObject{
				Id: connID2,
			},
			expected: expected{
				response: GetConnection200JSONResponse{
					Connection: Connection{
						CreatedAt: time.Now(),
						Id:        connID2.String(),
						IssuerID:  did.String(),
						UserID:    usrDID2.String(),
					},
					Credentials: common.ToPointer([]Credential{}),
				},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/connections/%s", tc.request.Id.String())

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetConnection200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				if tc.expected.response.Credentials != nil {
					require.NotNil(t, response.Credentials)
					require.Equal(t, len(*tc.expected.response.Credentials), len(*response.Credentials))
					for i := range *tc.expected.response.Credentials {
						validateCredential(t, (*tc.expected.response.Credentials)[i], (*response.Credentials)[i])
					}
				}
				assert.Equal(t, tc.expected.response.Connection.Id, response.Connection.Id)
				assert.Equal(t, tc.expected.response.Connection.IssuerID, response.Connection.IssuerID)
				assert.Equal(t, tc.expected.response.Connection.UserID, response.Connection.UserID)
				assert.InDelta(t, tc.expected.response.Connection.CreatedAt.Unix(), response.Connection.CreatedAt.Unix(), 10)
			case http.StatusBadRequest:
				var response GetConnection400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_GetConnections(t *testing.T) {
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
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaAdminMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	fixture := tests.NewFixture(storage)
	claim := fixture.NewClaim(t, did.String())
	fixture.CreateClaim(t, claim)

	usrDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	usrDID2, err := core.ParseDID("did:polygonid:polygon:mumbai:2qFBp1sRF1bFbTybVHHZQRgSWE2nKrdWeAxyZ67PdG")
	require.NoError(t, err)

	connID := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *did,
		UserDID:    *usrDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	connID2 := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *did,
		UserDID:    *usrDID2,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	handler := getHandler(ctx, server)

	type expected struct {
		response GetConnections200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		request  GetConnectionsRequestObject
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:    "No auth header",
			auth:    authWrong,
			request: GetConnectionsRequestObject{},
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "should return 2 connections",
			auth:    authOk,
			request: GetConnectionsRequestObject{},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Connection: Connection{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: time.Now(),
						},
					},
					{
						Connection: Connection{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: time.Now(),
						},
					},
				},
			},
		},
		{
			name: "should return 0 connections, no matching did",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("did:polygonid:polygon:mumbai:2qKZg1vCMwJeN4F5tyGhyjn8HPqHLJHS5eTWmud1Bj"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{},
			},
		},
		{
			name: "should return only one connection, full userDID",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer(usrDID.String()),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Connection: Connection{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: time.Now(),
						},
					},
				},
			},
		},
		{
			name: "should return only one connection, partial userDID",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Ge"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Connection: Connection{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: time.Now(),
						},
					},
				},
			},
		},
		{
			name: "should return two connections, beginning of did",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("did:"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Connection: Connection{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: time.Now(),
						},
					},
					{
						Connection: Connection{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: time.Now(),
						},
					},
				},
			},
		},
		{
			name: "should return two connections, invalid query for connections",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("some invalid did"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Connection: Connection{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: time.Now(),
						},
					},
					{
						Connection: Connection{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: time.Now(),
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			addr := "/v1/connections"
			if tc.request.Params.Query != nil {
				addr += "?query=" + *tc.request.Params.Query
			}

			req, err := http.NewRequest(http.MethodGet, addr, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetConnections200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				require.Equal(t, len(tc.expected.response), len(response))
				for i := range response {
					if tc.expected.response[i].Credentials != nil {
						require.NotNil(t, response[i].Credentials)
						require.Equal(t, len(*tc.expected.response[i].Credentials), len(*response[i].Credentials))
						for j := range *tc.expected.response[i].Credentials {
							validateCredential(t, (*tc.expected.response[i].Credentials)[j], (*response[i].Credentials)[j])
						}
					}
					assert.Equal(t, tc.expected.response[i].Connection.Id, response[i].Connection.Id)
					assert.Equal(t, tc.expected.response[i].Connection.IssuerID, response[i].Connection.IssuerID)
					assert.Equal(t, tc.expected.response[i].Connection.UserID, response[i].Connection.UserID)
					assert.InDelta(t, tc.expected.response[i].Connection.CreatedAt.Unix(), response[i].Connection.CreatedAt.Unix(), 10)
				}
			}
		})
	}
}

func validateCredential(t *testing.T, tc Credential, response Credential) {
	type credentialKYCSubject struct {
		Id           string `json:"id"`
		Birthday     uint64 `json:"birthday"`
		DocumentType uint64 `json:"documentType"`
		Type         string `json:"type"`
	}

	assert.Equal(t, tc.Id, response.Id)
	assert.Equal(t, tc.SchemaType, response.SchemaType)
	assert.Equal(t, tc.SchemaHash, response.SchemaHash)
	assert.Equal(t, tc.Revoked, response.Revoked)
	assert.Equal(t, tc.RevNonce, response.RevNonce)
	assert.InDelta(t, tc.CreatedAt.Unix(), response.CreatedAt.Unix(), 10)
	if response.ExpiresAt != nil && tc.ExpiresAt != nil {
		assert.InDelta(t, tc.ExpiresAt.Unix(), response.ExpiresAt.Unix(), 10)
	}
	assert.Equal(t, tc.Expired, response.Expired)
	var respAttributes, tcCredentialSubject credentialKYCSubject
	assert.NoError(t, mapstructure.Decode(tc.Attributes, &tcCredentialSubject))
	assert.NoError(t, mapstructure.Decode(response.Attributes, &respAttributes))
	assert.EqualValues(t, respAttributes, tcCredentialSubject)
	assert.EqualValues(t, tc.ProofTypes, response.ProofTypes)
}

func TestServer_RevokeCredential(t *testing.T) {
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
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)

	fixture := tests.NewFixture(storage)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewAdminSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	idClaim, err := uuid.NewUUID()
	require.NoError(t, err)
	nonce := int64(123)
	revNonce := domain.RevNonceUint64(nonce)
	fixture.CreateClaim(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      common.ToPointer(did.String()),
		Issuer:          did.String(),
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

	handler := getHandler(context.Background(), server)

	type expected struct {
		response RevokeCredentialResponseObject
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
			name:  "No auth header",
			auth:  authWrong,
			nonce: nonce,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:  "should revoke the claim",
			auth:  authOk,
			nonce: nonce,
			expected: expected{
				httpCode: 202,
				response: RevokeCredential202JSONResponse{
					Message: "claim revocation request sent",
				},
			},
		},
		{
			name:  "should get an error wrong nonce",
			auth:  authOk,
			nonce: int64(1231323),
			expected: expected{
				httpCode: 404,
				response: RevokeCredential404JSONResponse{N404JSONResponse{
					Message: "the claim does not exist",
				}},
			},
		},
		{
			name:  "should get an error - duplicated nonce",
			auth:  authOk,
			nonce: nonce,
			expected: expected{
				httpCode: 500,
				response: RevokeCredential500JSONResponse{N500JSONResponse{
					Message: "error revoking the claim: cannot add revocation nonce: 123 to revocation merkle tree: the entry index already exists in the tree",
				}},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/revoke/%d", tc.nonce)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch v := tc.expected.response.(type) {
			case RevokeCredential202JSONResponse:
				var response RevokeCredential202JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case RevokeCredential404JSONResponse:
				var response RevokeCredential404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			case RevokeCredential500JSONResponse:
				var response RevokeCredential500JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, response.Message, v.Message)
			}
		})
	}
}

func TestServer_CreateLink(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "mumbai"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
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
	linkRepository := repositories.NewLink(*storage)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil)
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(linkRepository)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaAdminSrv := services.NewSchemaAdmin(repositories.NewSchema(*storage), loader.HTTPFactory)
	importedSchema, err := schemaAdminSrv.ImportSchema(ctx, *did, url, schemaType)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewAdminSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateCredentialResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		body     CreateLinkRequest
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
			body: CreateLinkRequest{
				SchemaID: importedSchema.ID,
				ExpirationDate: &types.Date{
					Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local),
				},
				ClaimLinkExpiration: common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{{"birthday", "19790911"}, {"documentType", "12"}},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim link expiration invalid",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID: importedSchema.ID,
				ExpirationDate: &types.Date{
					Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local),
				},
				ClaimLinkExpiration: common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{{"birthday", "19790911"}, {"documentType", "12"}},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "invalid claimLinkExpiration. Cannot be a date time prior current time."}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link expiration nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID: importedSchema.ID,
				ExpirationDate: &types.Date{
					Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local),
				},
				ClaimLinkExpiration: nil,
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{{"birthday", "19790911"}, {"documentType", "12"}},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim expiration date nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:            importedSchema.ID,
				ExpirationDate:      nil,
				ClaimLinkExpiration: nil,
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{{"birthday", "19790911"}, {"documentType", "12"}},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim link wrong number of attributes",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID: importedSchema.ID,
				ExpirationDate: &types.Date{
					Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local),
				},
				ClaimLinkExpiration: common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "you must provide at least one attribute"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link wrong schema id",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID: uuid.New(),
				ExpirationDate: &types.Date{
					Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local),
				},
				ClaimLinkExpiration: common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				LimitedClaims:       common.ToPointer(10),
				Attributes:          []LinkRequestAttributesType{{"birthday", "19790911"}, {"documentType", "12"}},
				MtProof:             true,
				SignatureProof:      true,
			},
			expected: expected{
				response: CreateCredential400JSONResponse{N400JSONResponse{Message: "schema id not found"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/credentials/links"

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusCreated:
				var response UUIDResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				_, err := uuid.Parse(response.Id)
				assert.NoError(t, err)
			case http.StatusBadRequest:
				var response CreateCredential400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}
