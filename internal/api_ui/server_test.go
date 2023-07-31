package api_ui

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
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
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	schemaService := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)

	server := NewServer(&cfg, identityService, claimsService, schemaService, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), &health.Status{})
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
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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

	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, sessionRepository, pubsub.NewMock())
	server := NewServer(&cfg, identityService, NewClaimsMock(), NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	fixture := tests.NewFixture(storage)

	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/url",
		Type:      "schemaType",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
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

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	fixture := tests.NewFixture(storage)

	for i := 0; i < 20; i++ {
		s := &domain.Schema{
			ID:        uuid.New(),
			IssuerDID: *issuerDID,
			URL:       fmt.Sprintf("https://domain.org/this/is/an/url/%d", i),
			Type:      fmt.Sprintf("schemaType-%d", i),
			Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
			CreatedAt: time.Now(),
		}
		s.Hash = utils.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
		fixture.CreateSchema(t, ctx, s)
	}
	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *issuerDID,
		URL:       "https://domain.org/this/is/an/url/ubiprogram",
		Type:      "UbiProgram",
		Words:     domain.SchemaWordsFromString("attr1, attr2, attr3"),
		CreatedAt: time.Now(),
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
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
				SchemaType:  "lala",
				Url:         "wrong/url",
				Title:       common.ToPointer("some Title"),
				Description: common.ToPointer("some Description"),
				Version:     uuid.NewString(),
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
				SchemaType:  schemaType,
				Url:         url,
				Title:       common.ToPointer("some Title"),
				Description: common.ToPointer("some Description"),
				Version:     uuid.NewString(),
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

func TestServer_ImportSchemaIPFS(t *testing.T) {
	// TIP: A copy of the files here internal/api_ui/testdata/ipfs-schema-1.json
	const url = "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL"
	const schemaType = "testNewType"
	ctx := context.Background()
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
				SchemaType:  "lala",
				Url:         "wrong/url",
				Title:       common.ToPointer("title"),
				Description: common.ToPointer("description"),
				Version:     "1.0.0",
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
				SchemaType:  schemaType,
				Url:         url,
				Title:       common.ToPointer("title"),
				Description: common.ToPointer("description"),
				Version:     "1.0.0",
			},
			expected: expected{
				httpCode: http.StatusCreated,
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	issuerDID, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	server.cfg.APIUI.IssuerDID = *issuerDID
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	userDID2, err := core.ParseDID("did:polygonid:polygon:mumbai:2qNytPv6dKKhfqopjBdXJU1vSVb3Lbgcidved32R64")
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
		RevNonce:        0,
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
			urlTest := fmt.Sprintf("/v1/connections/%s", tc.connID.String())
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
			req, err := http.NewRequest("DELETE", parsedURL.String(), nil)
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
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	issuerDID, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	pubSub := pubsub.NewMock()
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubSub, ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response                    CreateCredentialResponseObject
		httpCode                    int
		createCredentialEventsCount int
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
				Expiration:     common.ToPointer(time.Now()),
				SignatureProof: common.ToPointer(true),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
			},
		},
		{
			name: "Happy path with IPFS schema",
			auth: authOk,
			body: CreateCredentialRequest{
				CredentialSchema: "ipfs://QmQVeb5dkz5ekDqBrYVVxBFQZoCbzamnmMUn9B8twCEgDL",
				Type:             "testNewType",
				CredentialSubject: map[string]any{
					"id":             "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
					"testNewTypeInt": 1,
				},
				Expiration:     common.ToPointer(time.Now()),
				SignatureProof: common.ToPointer(true),
			},
			expected: expected{
				response:                    CreateCredential201JSONResponse{},
				httpCode:                    http.StatusCreated,
				createCredentialEventsCount: 1,
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
				Expiration: common.ToPointer(time.Now()),
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
				Expiration:     common.ToPointer(time.Now()),
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
				Expiration:     common.ToPointer(time.Now()),
				SignatureProof: common.ToPointer(true),
			},
			expected: expected{
				response: CreateCredential422JSONResponse{N422JSONResponse{Message: "cannot load schema"}},
				httpCode: http.StatusUnprocessableEntity,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pubSub.Clear(event.CreateCredentialEvent)

			rr := httptest.NewRecorder()
			url := "/v1/credentials"

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			assert.Equal(t, tc.expected.createCredentialEventsCount, len(pubSub.AllPublishedEvents(event.CreateCredentialEvent)))

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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
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
			name:         "should get an error, a credential cannot be deleted twice",
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	createdClaim1, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)

	createdClaim2, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(false), nil, false))
	require.NoError(t, err)

	createdClaim3, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(false), common.ToPointer(true), nil, false))
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
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim1.ID,
					ProofTypes: []string{"BJJSignature2021", "SparseMerkleTreeProof"},
					RevNonce:   uint64(createdClaim1.RevNonce),
					Revoked:    createdClaim1.Revoked,
					SchemaHash: createdClaim1.SchemaHash,
					SchemaType: typeC,
					SchemaUrl:  schema,
					UserID:     createdClaim1.OtherIdentifier,
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
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim2.ID,
					ProofTypes: []string{"BJJSignature2021"},
					RevNonce:   uint64(createdClaim2.RevNonce),
					Revoked:    createdClaim2.Revoked,
					SchemaHash: createdClaim2.SchemaHash,
					SchemaType: typeC,
					SchemaUrl:  schema,
					UserID:     createdClaim2.OtherIdentifier,
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
					CredentialSubject: map[string]interface{}{
						"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
						"birthday":     19960424,
						"documentType": 2,
					},
					CreatedAt:  time.Now().UTC(),
					Expired:    false,
					ExpiresAt:  nil,
					Id:         createdClaim3.ID,
					ProofTypes: []string{"SparseMerkleTreeProof"},
					RevNonce:   uint64(createdClaim3.RevNonce),
					Revoked:    createdClaim3.Revoked,
					SchemaHash: createdClaim3.SchemaHash,
					SchemaType: typeC,
					SchemaUrl:  schema,
					UserID:     createdClaim3.OtherIdentifier,
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
	schemaRepository := repositories.NewSchema(*storage)
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	schemaService := services.NewSchema(schemaRepository, schemaLoader)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schemaURL := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	future := time.Now().Add(1000 * time.Hour)
	past := time.Now().Add(-1000 * time.Hour)
	iReq := ports.NewImportSchemaRequest(schemaURL, typeC, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	_, err = schemaService.ImportSchema(ctx, *did, iReq)
	require.NoError(t, err)
	// Never expires
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)

	// Expires in future
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(false), nil, false))
	require.NoError(t, err)

	// Expired
	claim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject, &past, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(false), nil, false))
	require.NoError(t, err)

	// non expired, but revoked
	revoked, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(false), common.ToPointer(true), nil, false))
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
		did      *string
		query    *string
		status   *string
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
			name:   "Wrong status",
			auth:   authOk,
			status: common.ToPointer("wrong"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "wrong type value. Allowed values: [all, revoked, expired]",
			},
		},
		{
			name: "wrong did",
			auth: authOk,
			did:  common.ToPointer("wrongdid:"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				errorMsg: "cannot parse did parameter: wrong format",
			},
		},
		{
			name: "Get all implicit",
			auth: authOk,
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:   "Get all explicit",
			auth:   authOk,
			status: common.ToPointer("all"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:   "Get all from existing did",
			auth:   authOk,
			status: common.ToPointer("all"),
			did:    &claim.OtherIdentifier,
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:   "Get all from non existing did. Expecting empty list",
			auth:   authOk,
			status: common.ToPointer("all"),
			did:    common.ToPointer("did:iden3:tJU7z1dbKyKYLiaopZ5tN6Zjsspq7QhYayiR31RFa"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    0,
			},
		},
		{
			name:   "Revoked",
			auth:   authOk,
			status: common.ToPointer("revoked"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:   "REVOKED",
			auth:   authOk,
			status: common.ToPointer("REVOKED"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:   "Expired",
			auth:   authOk,
			status: common.ToPointer("expired"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    1,
			},
		},
		{
			name:  "Search by did and other words in query params:",
			auth:  authOk,
			query: common.ToPointer("some words and " + revoked.OtherIdentifier),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "Search by partial did and other words in query params:",
			auth:  authOk,
			query: common.ToPointer("some words and " + revoked.OtherIdentifier[9:14]),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "Search by did in query params:",
			auth:  authOk,
			query: &revoked.OtherIdentifier,
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "Search by attributes in query params",
			auth:  authOk,
			query: common.ToPointer("birthday"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "Search by attributes in query params, partial word",
			auth:  authOk,
			query: common.ToPointer("rthd"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "Search by partial did in query params:",
			auth:  authOk,
			query: common.ToPointer(revoked.OtherIdentifier[9:14]),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "FTS is doing and OR when no did passed:",
			auth:  authOk,
			query: common.ToPointer("birthday schema attribute not the rest of words this sentence"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    4,
			},
		},
		{
			name:  "FTS is doing and AND when did passed:",
			auth:  authOk,
			did:   &claim.OtherIdentifier,
			query: common.ToPointer("not existing words"),
			expected: expected{
				httpCode: http.StatusOK,
				count:    0,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := url.URL{Path: "/v1/credentials"}
			queryParams := make([]string, 0)
			if tc.query != nil {
				queryParams = append(queryParams, "query="+*tc.query)
			}
			if tc.status != nil {
				queryParams = append(queryParams, "status="+*tc.status)
			}
			if tc.did != nil {
				queryParams = append(queryParams, "did="+*tc.did)
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
				assert.Len(t, response, tc.expected.count)
			case http.StatusBadRequest:
				var response GetCredentials400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.errorMsg, response.Message)
			}
		})
	}
}

func TestServer_GetCredentialQrCode(t *testing.T) {
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(ctx, server)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	createdClaim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)

	type expected struct {
		message  *string
		response QrCodeResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		request  GetCredentialRequestObject
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "should return an error, claim not found",
			request: GetCredentialRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				message:  common.ToPointer("Credential not found"),
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "happy path",
			request: GetCredentialRequestObject{
				Id: createdClaim.ID,
			},
			expected: expected{
				response: QrCodeResponse{
					Body: QrCodeBodyResponse{
						Credentials: []QrCodeCredentialResponse{
							{
								Description: schema,
								Id:          createdClaim.ID.String(),
							},
						},
						Url: "",
					},
					From: did.String(),
					To:   createdClaim.OtherIdentifier,
				},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/%s/qrcode", tc.request.Id.String())

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response QrCodeResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				require.Equal(t, tc.expected.response.From, response.From)
				require.Equal(t, tc.expected.response.To, response.To)
				require.Equal(t, len(tc.expected.response.Body.Credentials), len(response.Body.Credentials))
			case http.StatusBadRequest:
				var response GetCredential400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, *tc.expected.message, response.Message)
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

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
					CreatedAt: time.Now(),
					Id:        connID.String(),
					IssuerID:  did.String(),
					UserID:    usrDID.String(),
					Credentials: []Credential{
						{
							CredentialSubject: map[string]interface{}{
								"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
								"birthday":     19960424,
								"documentType": 2,
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
							SchemaUrl:  claim.SchemaURL,
							UserID:     claim.OtherIdentifier,
						},
					},
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
					CreatedAt:   time.Now(),
					Id:          connID2.String(),
					IssuerID:    did.String(),
					UserID:      usrDID2.String(),
					Credentials: []Credential{},
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
					require.Equal(t, len(tc.expected.response.Credentials), len(response.Credentials))
					for i := range tc.expected.response.Credentials {
						validateCredential(t, (tc.expected.response.Credentials)[i], (response.Credentials)[i])
					}
				}
				assert.Equal(t, tc.expected.response.Id, response.Id)
				assert.Equal(t, tc.expected.response.IssuerID, response.IssuerID)
				assert.Equal(t, tc.expected.response.UserID, response.UserID)
				assert.InDelta(t, tc.expected.response.CreatedAt.Unix(), response.CreatedAt.Unix(), 10)
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)

	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	fixture := tests.NewFixture(storage)

	schemaURL := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	schemaContext := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
	schemaType := "KYCAgeCredential"
	s := &domain.Schema{
		ID:        uuid.New(),
		IssuerDID: *did,
		URL:       schemaURL,
		Type:      schemaType,
		Words:     []string{"birthday", "id", "hello"},
		CreatedAt: time.Now(),
		Hash:      utils.CreateSchemaHash([]byte(schemaContext + "#" + schemaType)),
	}
	fixture.CreateSchema(t, ctx, s)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	credentialSubject2 := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960423,
		"documentType": 2,
	}

	merklizedRootPosition := "index"
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject, nil, schemaType, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schemaURL, credentialSubject2, nil, schemaType, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)

	usrDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	usrDID2, err := core.ParseDID("did:polygonid:polygon:mumbai:2qFBp1sRF1bFbTybVHHZQRgSWE2nKrdWeAxyZ67PdG")
	require.NoError(t, err)

	uuid1, err := uuid.Parse("9736cf94-cd42-11ed-9618-debe37e1cbd6")
	require.NoError(t, err)
	connID := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid1,
		IssuerDID:  *did,
		UserDID:    *usrDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	uuid2, err := uuid.Parse("5736cf94-cd42-11ed-9618-debe37e1cbd6")
	require.NoError(t, err)
	connID2 := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid2,
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
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
						CreatedAt: time.Now(),
					},
					{
						Id:        connID2.String(),
						IssuerID:  did.String(),
						UserID:    usrDID2.String(),
						CreatedAt: time.Now(),
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
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
						CreatedAt: time.Now(),
					},
				},
			},
		},
		{
			name: "should return only one connection, partial userDID",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("Z7gcmEoP2KppvFPCZqyzyb5tK9T6Ge"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
						CreatedAt: time.Now(),
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
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
						CreatedAt: time.Now(),
					},
					{
						Id:        connID2.String(),
						IssuerID:  did.String(),
						UserID:    usrDID2.String(),
						CreatedAt: time.Now(),
					},
				},
			},
		},
		{
			name: "should return two connections, one of it with credentials",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:          connID.String(),
						IssuerID:    did.String(),
						UserID:      usrDID.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{{}, {}},
					},
					{
						Id:          connID2.String(),
						IssuerID:    did.String(),
						UserID:      usrDID2.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{},
					},
				},
			},
		},
		{
			name: "should return one connection with credentials",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("5HFANQ"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:          connID.String(),
						IssuerID:    did.String(),
						UserID:      usrDID.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{{}, {}},
					},
				},
			},
		},
		{
			name: "should return one connection with credentials partial did",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9 "),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:          connID.String(),
						IssuerID:    did.String(),
						UserID:      usrDID.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{{}, {}},
					},
				},
			},
		},
		{
			name: "should return one connection with credentials partial did and attributes",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("CZqyzyb5tK9T6Ge  credential"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:          connID.String(),
						IssuerID:    did.String(),
						UserID:      usrDID.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{{}, {}},
					},
				},
			},
		},
		{
			name: "should return one connection with not existing did and valid attributes",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("did:polygon:myhouse:ZZZZZZ birthday"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					{
						Id:          connID.String(),
						IssuerID:    did.String(),
						UserID:      usrDID.String(),
						CreatedAt:   time.Now(),
						Credentials: []Credential{{}, {}},
					},
				},
			},
		},
		{
			name: "should return 0 connections with invalid did and invalid attributes",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("did:polygonid:polygon:mumbai:2qFVUasb8QZ1XAmD71b3NA8bzQhGs92VQEPgELYnpk nothingValid here"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{},
			},
		},
		{
			name: "should return no connections, did not found",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("did:polygonid:polygon:mumbai:2qFVUasb8QZ1XAmD71b3NA8bzQhGs92VQEPgELYnpk"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{},
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
						Id:        connID.String(),
						IssuerID:  did.String(),
						UserID:    usrDID.String(),
						CreatedAt: time.Now(),
					},
					{
						Id:        connID2.String(),
						IssuerID:  did.String(),
						UserID:    usrDID2.String(),
						CreatedAt: time.Now(),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			addr := "/v1/connections"

			parsedURL, err := url.Parse(addr)
			require.NoError(t, err)
			values := parsedURL.Query()
			if tc.request.Params.Query != nil {
				values.Add("query", *tc.request.Params.Query)
			}
			if tc.request.Params.Credentials != nil && *tc.request.Params.Credentials {
				values.Add("credentials", "true")
			}
			parsedURL.RawQuery = values.Encode()
			req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
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
						require.Equal(t, len(tc.expected.response[i].Credentials), len(response[i].Credentials))
					}
					assert.Equal(t, tc.expected.response[i].Id, response[i].Id)
					assert.Equal(t, tc.expected.response[i].IssuerID, response[i].IssuerID)
					assert.Equal(t, tc.expected.response[i].UserID, response[i].UserID)
					assert.InDelta(t, tc.expected.response[i].CreatedAt.Unix(), response[i].CreatedAt.Unix(), 10)
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
	assert.NoError(t, mapstructure.Decode(tc.CredentialSubject, &tcCredentialSubject))
	assert.NoError(t, mapstructure.Decode(response.CredentialSubject, &respAttributes))
	assert.EqualValues(t, respAttributes, tcCredentialSubject)
	assert.EqualValues(t, tc.ProofTypes, response.ProofTypes)
	assert.Equal(t, tc.UserID, response.UserID)
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
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)

	fixture := tests.NewFixture(storage)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

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
	schemaRespository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	pubSub := pubsub.NewMock()
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubSub, ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRespository, loader.HTTPFactory, sessionRepository, pubSub, ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response CreateLinkResponseObject
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
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "No merkle tree proof or signature proof selected. At least one should be enabled",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              false,
				SignatureProof:       false,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "at least one proof type should be enabled"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link expiration exceeded",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "invalid claimLinkExpiration. Cannot be a date time prior current time."}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link expiration nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: nil,
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim expiration date nil",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           nil,
				CredentialExpiration: nil,
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink201JSONResponse{},
				httpCode: http.StatusCreated,
			},
		},
		{
			name: "Claim link wrong number of attributes",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "you must provide at least one attribute"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link wrong attribute type",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             importedSchema.ID,
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": true},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "cannot parse claim"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link wrong schema id",
			auth: authOk,
			body: CreateLinkRequest{
				SchemaID:             uuid.New(),
				Expiration:           common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
				CredentialExpiration: &types.Date{Time: time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)},
				LimitedClaims:        common.ToPointer(10),
				CredentialSubject:    CredentialSubject{"birthday": 19790911, "documentType": 12},
				MtProof:              true,
				SignatureProof:       true,
			},
			expected: expected{
				response: CreateLink400JSONResponse{N400JSONResponse{Message: "schema does not exist"}},
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
				var response CreateLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_ActivateLink(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	tomorrow := time.Now().Add(24 * time.Hour)
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, nil, true, true, CredentialSubject{"birthday": 19790911, "documentType": 12})
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response AcivateLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		body     AcivateLinkJSONBody
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			body: AcivateLinkJSONBody{
				Active: true,
			},
			expected: expected{
				response: AcivateLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Claim link already activated",
			auth: authOk,
			id:   link.ID,
			body: AcivateLinkJSONBody{
				Active: true,
			},
			expected: expected{
				response: AcivateLink400JSONResponse{N400JSONResponse{Message: "link is already active"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			id:   link.ID,
			body: AcivateLinkJSONBody{
				Active: false,
			},
			expected: expected{
				response: AcivateLink200JSONResponse{Message: "Link updated"},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "Claim link already deactivated",
			auth: authOk,
			id:   link.ID,
			body: AcivateLinkJSONBody{
				Active: false,
			},
			expected: expected{
				response: AcivateLink400JSONResponse{N400JSONResponse{Message: "link is already inactive"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s", tc.id)

			req, err := http.NewRequest(http.MethodPatch, url, tests.JSONBody(t, tc.body))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(AcivateLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response AcivateLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

// TestServer_GetLink does an end 2 end test for the get link endpoint.
// TIP: Link status test is better covered in the internal/repositories/tests/link_test.go unit tests
// as it is really verbose to do it here.
func TestServer_GetLink(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	require.NoError(t, err)
	hash, _ := link.Schema.Hash.MarshalText()

	linkExpired, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				response: GetLink404JSONResponse{N404JSONResponse{Message: "link not found"}},
				httpCode: http.StatusNotFound,
			},
		},
		{
			name: "Happy path, link active by date",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLink200JSONResponse{
					Active:            link.Active,
					CredentialSubject: CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:        link.ValidUntil,
					Id:                link.ID,
					IssuedClaims:      link.IssuedClaims,
					MaxIssuance:       link.MaxIssuance,
					SchemaType:        link.Schema.Type,
					SchemaUrl:         link.Schema.URL,
					Status:            LinkStatusActive,
					ProofTypes:        []string{"SparseMerkleTreeProof", "BJJSignature2021"},
					CreatedAt:         link.CreatedAt,
					SchemaHash:        string(hash),
				},
			},
		},
		{
			name: "Happy path, link expired by date",
			auth: authOk,
			id:   linkExpired.ID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLink200JSONResponse{
					Active:            linkExpired.Active,
					CredentialSubject: CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:        linkExpired.ValidUntil,
					Id:                linkExpired.ID,
					IssuedClaims:      linkExpired.IssuedClaims,
					MaxIssuance:       linkExpired.MaxIssuance,
					SchemaType:        linkExpired.Schema.Type,
					SchemaUrl:         linkExpired.Schema.URL,
					Status:            LinkStatusExceeded,
					ProofTypes:        []string{"SparseMerkleTreeProof", "BJJSignature2021"},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s", tc.id)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetLink200JSONResponse
				expected, ok := tc.expected.response.(GetLink200JSONResponse)
				require.True(t, ok)
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, expected.Active, response.Active)
				assert.Equal(t, expected.MaxIssuance, response.MaxIssuance)
				assert.Equal(t, expected.Status, response.Status)
				assert.Equal(t, expected.IssuedClaims, response.IssuedClaims)
				assert.Equal(t, expected.Id, response.Id)
				tcCred, err := json.Marshal(expected.CredentialSubject)
				require.NoError(t, err)
				respCred, err := json.Marshal(response.CredentialSubject)
				require.NoError(t, err)
				assert.Equal(t, tcCred, respCred)
				assert.Equal(t, expected.SchemaType, response.SchemaType)
				assert.Equal(t, expected.SchemaUrl, response.SchemaUrl)
				assert.Equal(t, expected.Active, response.Active)
				assert.InDelta(t, expected.Expiration.UnixMilli(), response.Expiration.UnixMilli(), 10)
				assert.Equal(t, len(expected.ProofTypes), len(response.ProofTypes))
			case http.StatusNotFound:
				var response GetLink404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(GetLink404JSONResponse)
				require.True(t, ok)
				assert.EqualValues(t, expected.Message, response.Message)
			}
		})
	}
}

func TestServer_GetAllLinks(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "mumbai"
		sUrl       = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(sUrl, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link1, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	require.NoError(t, err)
	linkActive := getLinkResponse(*link1)

	time.Sleep(10 * time.Millisecond)

	link2, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	require.NoError(t, err)
	linkExpired := getLinkResponse(*link2)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	link3, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	link3.Active = false
	require.NoError(t, err)
	require.NoError(t, linkService.Activate(ctx, *did, link3.ID, false))
	linkInactive := getLinkResponse(*link3)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	handler := getHandler(ctx, server)
	type expected struct {
		response []Link
		httpCode int
	}
	type testConfig struct {
		name     string
		query    *string
		status   *GetLinksParamsStatus
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
			name:   "Wrong type",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("unknown-filter-type")),
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},

		{
			name: "Happy path. All schemas",
			auth: authOk,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired, linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, explicit",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("all")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired, linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, active",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("active")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkActive},
			},
		},
		{
			name:   "Happy path. All schemas, exceeded",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Happy path. All schemas, exceeded",
			auth:   authOk,
			status: common.ToPointer(GetLinksParamsStatus("inactive")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive},
			},
		},
		{
			name:   "Exceeded with filter found",
			auth:   authOk,
			query:  common.ToPointer("documentType"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Exceeded with filter found, partial match",
			auth:   authOk,
			query:  common.ToPointer("docum"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
		{
			name:   "Empty result",
			auth:   authOk,
			query:  common.ToPointer("documentType"),
			status: common.ToPointer(GetLinksParamsStatus("exceeded")),
			expected: expected{
				httpCode: http.StatusOK,
				response: GetLinks200JSONResponse{linkInactive, linkExpired},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			endpoint := url.URL{Path: "/v1/credentials/links"}
			if tc.status != nil {
				endpoint.RawQuery = endpoint.RawQuery + "status=" + string(*tc.status)
			}
			if tc.query != nil {
				endpoint.RawQuery = endpoint.RawQuery + "&query=" + *tc.query
			}

			req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetLinks200JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				if assert.Equal(t, len(tc.expected.response), len(response)) {
					for i, resp := range response {
						assert.Equal(t, tc.expected.response[i].Id, resp.Id)
						assert.Equal(t, tc.expected.response[i].Status, resp.Status)
						assert.Equal(t, tc.expected.response[i].IssuedClaims, resp.IssuedClaims)
						assert.Equal(t, tc.expected.response[i].Active, resp.Active)
						assert.Equal(t, tc.expected.response[i].MaxIssuance, resp.MaxIssuance)
						assert.Equal(t, tc.expected.response[i].SchemaUrl, resp.SchemaUrl)
						assert.Equal(t, tc.expected.response[i].SchemaType, resp.SchemaType)
						tcCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						respCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						assert.Equal(t, tcCred, respCred)
						assert.InDelta(t, tc.expected.response[i].Expiration.UnixMilli(), resp.Expiration.UnixMilli(), 10)
					}
				}
			case http.StatusBadRequest:
				var response GetLinks400JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
			}
		})
	}
}

func TestServer_DeleteLink(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				response: DeleteLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "Happy path",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				response: DeleteLink200JSONResponse{Message: "link deleted"},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s", tc.id)

			req, err := http.NewRequest(http.MethodDelete, url, tests.JSONBody(t, nil))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(DeleteLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response DeleteLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_DeleteLinkForDifferentDID(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	iden2, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	did2, err := core.ParseDID(iden2.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did2
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	type expected struct {
		response DeleteLinkResponseObject
		httpCode int
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		auth     func() (string, string)
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "No auth header",
			auth: authWrong,
			id:   link.ID,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Claim link does not exist for a did",
			auth: authOk,
			id:   link.ID,
			expected: expected{
				response: DeleteLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}},
				httpCode: http.StatusBadRequest,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s", tc.id)

			req, err := http.NewRequest(http.MethodDelete, url, tests.JSONBody(t, nil))
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GenericMessage
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				expected, ok := tc.expected.response.(DeleteLink200JSONResponse)
				assert.True(t, ok)
				assert.Equal(t, expected.Message, response.Message)

			case http.StatusBadRequest:
				var response DeleteLink400JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.response, response)
			}
		})
	}
}

func TestServer_CreateLinkQRCode(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	cfg.APIUI.ServerURL = "http://localhost/issuer-admin"

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 0, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 0, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)

	yesterday := time.Now().Add(-24 * time.Hour)
	linkExpired, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	require.NoError(t, err)

	handler := getHandler(ctx, server)

	linkDetail := getLinkResponse(*link)

	type expected struct {
		linkDetail Link
		httpCode   int
		message    string
	}

	type testConfig struct {
		name     string
		id       uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "Wrong link id",
			id:   uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: link not found",
			},
		},
		{
			name: "Expired link",
			id:   linkExpired.ID,
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: cannot issue a credential for an expired link",
			},
		},
		{
			name: "Happy path",
			id:   link.ID,
			expected: expected{
				linkDetail: linkDetail,
				httpCode:   http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s/qrcode", tc.id.String())

			req, err := http.NewRequest(http.MethodPost, url, tests.JSONBody(t, nil))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				callBack := cfg.APIUI.ServerURL + "/v1/credentials/links/callback?"
				var response CreateLinkQrCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.NotNil(t, response.QrCode.Body)
				assert.Equal(t, "authentication", response.QrCode.Body.Reason)
				callbackArr := strings.Split(response.QrCode.Body.CallbackUrl, "sessionID")
				assert.True(t, len(callbackArr) == 2)
				assert.Equal(t, callBack, callbackArr[0])
				params := strings.Split(callbackArr[1], "linkID")
				assert.True(t, len(params) == 2)
				assert.NotNil(t, response.QrCode.Id)
				assert.Equal(t, "https://iden3-communication.io/authorization/1.0/request", response.QrCode.Type)
				assert.Equal(t, "application/iden3comm-plain-json", response.QrCode.Typ)
				assert.Equal(t, cfg.APIUI.IssuerDID.String(), response.QrCode.From)
				assert.NotNil(t, response.QrCode.Thid)
				assert.NotNil(t, response.SessionID)
				assert.Equal(t, tc.expected.linkDetail.Id, response.LinkDetail.Id)
				assert.Equal(t, tc.expected.linkDetail.SchemaType, response.LinkDetail.SchemaType)
			case http.StatusNotFound:
				var response CreateLinkQrCode404JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.EqualValues(t, tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_GetLinkQRCode(t *testing.T) {
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
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.HTTPFactory, cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, loader.HTTPFactory, sessionRepository, pubsub.NewMock(), ipfsGateway)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), loader.HTTPFactory)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	cfg.APIUI.ServerURL = "http://localhost/issuer-admin"

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, NewPublisherMock(), NewPackageManagerMock(), nil)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 0, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 0, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	sessionID := uuid.New()
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qP8KN3KRwBi37jB2ENXrWxhTo3pefaU5u5BFPbjYo")
	assert.NoError(t, err)
	qrcode := &linkState.QRCodeMessage{
		ID:       uuid.New().String(),
		Typ:      "application/iden3comm-plain-json",
		Type:     "https://iden3-communication.io/credentials/1.0/offer",
		ThreadID: uuid.New().String(),
		Body: linkState.CredentialsLinkMessageBody{
			URL: "https://domain/issuer/v1/agent",
			Credentials: []linkState.CredentialLink{
				{
					ID:          uuid.NewString(),
					Description: "KYCAgeCredential",
				},
			},
		},
		From: did.String(),
		To:   userDID.String(),
	}

	linkDetail := getLinkResponse(*link)

	type expected struct {
		qrCode     *linkState.QRCodeMessage
		linkDetail Link
		status     string
		httpCode   int
	}

	type testConfig struct {
		name      string
		id        uuid.UUID
		sessionID uuid.UUID
		state     *linkState.State
		expected  expected
	}

	for _, tc := range []testConfig{
		{
			name:      "Wrong sessionID",
			sessionID: uuid.New(),
			id:        link.ID,
			state:     nil,
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:      "Wrong linkID",
			sessionID: sessionID,
			id:        uuid.New(),
			state:     nil,
			expected: expected{
				httpCode: http.StatusNotFound,
			},
		},
		{
			name:      "Error state",
			sessionID: sessionID,
			id:        link.ID,
			state:     linkState.NewStateError(errors.New("something wrong")),
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name:      "Pending state",
			sessionID: sessionID,
			id:        link.ID,
			state:     linkState.NewStatePending(),
			expected: expected{
				httpCode: http.StatusOK,
				status:   "pending",
			},
		},
		{
			name:      "Happy path",
			id:        link.ID,
			sessionID: sessionID,
			state:     linkState.NewStateDone(qrcode),
			expected: expected{
				linkDetail: linkDetail,
				qrCode:     qrcode,
				status:     "done",
				httpCode:   http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.state != nil {
				err := sessionRepository.SetLink(ctx, linkState.CredentialStateCacheKey(tc.id.String(), tc.sessionID.String()), *tc.state)
				assert.NoError(t, err)
			}

			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/links/%s/qrcode?sessionID=%s", tc.id, tc.sessionID)
			req, err := http.NewRequest(http.MethodGet, url, tests.JSONBody(t, nil))
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetLinkQRCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				if tc.expected.status == "pending" {
					assert.Equal(t, tc.expected.status, *response.Status)
				} else {
					assert.NotNil(t, response.QrCode.Body)
					assert.Equal(t, tc.expected.qrCode.Body.Credentials[0].ID, response.QrCode.Body.Credentials[0].Id)
					assert.Equal(t, tc.expected.qrCode.Body.Credentials[0].Description, response.QrCode.Body.Credentials[0].Description)
					assert.Equal(t, tc.expected.qrCode.Body.URL, response.QrCode.Body.Url)
					assert.NotNil(t, response.QrCode.Id)
					assert.Equal(t, tc.expected.qrCode.Type, response.QrCode.Type)
					assert.Equal(t, tc.expected.qrCode.Typ, response.QrCode.Typ)
					assert.Equal(t, tc.expected.qrCode.From, response.QrCode.From)
					assert.Equal(t, tc.expected.linkDetail.Id, response.LinkDetail.Id)
					assert.Equal(t, tc.expected.linkDetail.SchemaType, response.LinkDetail.SchemaType)
					assert.Equal(t, tc.expected.status, *response.Status)
				}
			}
		})
	}
}

func TestServer_GetStateStatus(t *testing.T) {
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetStateStatus200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		auth     func() (string, string)
		cleanUp  func()
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
			name: "No states to process",
			auth: authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: false},
				httpCode: http.StatusOK,
			},
			cleanUp: func() {
				cred, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, true))
				require.NoError(t, err)
				require.NoError(t, claimsService.Revoke(ctx, cfg.APIUI.IssuerDID, uint64(cred.RevNonce), "not valid"))
			},
		},
		{
			name: "New state to process",
			auth: authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: true},
				httpCode: http.StatusOK,
			},
			cleanUp: func() {},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/state/status"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetStateStatus200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.PendingActions, response.PendingActions)
				tc.cleanUp()
			}
		})
	}
}

func TestServer_GetStateTransactions(t *testing.T) {
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	handler := getHandler(ctx, server)

	type expected struct {
		response GetStateTransactions200JSONResponse
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
			name: "No states to process",
			auth: authOk,
			expected: expected{
				response: GetStateTransactions200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
		{
			name: "No state transactions after revoking/creating credentials",
			auth: authOk,
			expected: expected{
				response: GetStateTransactions200JSONResponse{},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/state/transactions"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetStateTransactions200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, len(tc.expected.response), len(response))
			}
		})
	}
}

func TestServer_GetRevocationStatus(t *testing.T) {
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory(ipfsGateway), cachex)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "http://host",
	}
	pubSub := pubsub.NewMock()
	claimsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf, pubSub, ipfsGateway)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), NewPublisherMock(), NewPackageManagerMock(), nil)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	createdCredential, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	require.NoError(t, err)

	handler := getHandler(ctx, server)

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
			nonce: int64(createdCredential.RevNonce),
			expected: expected{
				httpCode: http.StatusOK,
			},
		},

		{
			name:  "should get revocation status wrong nonce",
			nonce: 123456789,
			expected: expected{
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/credentials/revocation/status/%d", tc.nonce)
			req, err := http.NewRequest("GET", url, nil)
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
