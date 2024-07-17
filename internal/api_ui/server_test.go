package api_ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
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
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
	networkPkg "github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func TestServer_CheckStatus(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	schemaService := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, "http://localhost", pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	server := NewServer(&cfg, identityService, claimsService, schemaService, NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), &health.Status{}, *networkResolver)
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
	ctx := context.Background()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
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

func TestServer_GetAuthenticationConnection(t *testing.T) {
	ctx := context.Background()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	connectionRepository := repositories.NewConnections()
	claimsRepository := repositories.NewClaims()
	qrService := services.NewQrStoreService(cachex)
	connectionsService := services.NewConnection(connectionRepository, claimsRepository, storage)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), connectionsService, NewLinkMock(), qrService, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qKDJmySKNi4GD4vYdqfLb37MSTSijg77NoRZaKfDX")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)
	conn := &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	fixture.CreateConnection(t, conn)
	require.NoError(t, err)

	sessionID := uuid.New()
	fixture.CreateUserAuthentication(t, conn.ID, sessionID, conn.CreatedAt)

	type expected struct {
		httpCode int
		message  string
		response GetAuthenticationConnection200JSONResponse
	}
	type testConfig struct {
		name     string
		auth     func() (string, string)
		id       uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name: "Not authorized",
			auth: authWrong,
			id:   uuid.New(),
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name: "Session Not found",
			auth: authOk,
			id:   uuid.New(),
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  services.ErrConnectionDoesNotExist.Error(),
			},
		},
		{
			name: "Happy path. Existing connection",
			auth: authOk,
			id:   sessionID,
			expected: expected{
				httpCode: http.StatusOK,
				response: GetAuthenticationConnection200JSONResponse{
					Connection: AuthenticationConnection{
						Id:         conn.ID.String(),
						IssuerID:   conn.IssuerDID.String(),
						CreatedAt:  TimeUTC(conn.CreatedAt),
						ModifiedAt: TimeUTC(conn.ModifiedAt),
						UserID:     conn.UserDID.String(),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := fmt.Sprintf("/v1/authentication/sessions/%s", tc.id.String())
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)
			req.SetBasicAuth(tc.auth())

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetAuthenticationConnection200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.Connection.Id, response.Connection.Id)
				assert.Equal(t, tc.expected.response.Connection.IssuerID, response.Connection.IssuerID)
				assert.InDelta(t, time.Time(tc.expected.response.Connection.CreatedAt).Unix(), time.Time(response.Connection.CreatedAt).Unix(), 100)
				assert.InDelta(t, time.Time(tc.expected.response.Connection.ModifiedAt).Unix(), time.Time(response.Connection.ModifiedAt).Unix(), 100)
				assert.Equal(t, tc.expected.response.Connection.UserID, response.Connection.UserID)
			case http.StatusNotFound:
				var response GetAuthenticationConnection404JSONResponse
				assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.message, response.Message)
			}
		})
	}
}

func TestServer_AuthQRCode(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	sessionRepository := repositories.NewSessionCached(cachex)

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, sessionRepository, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	server := NewServer(&cfg, identityService, NewClaimsMock(), NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), qrService, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = *issuerDID
	server.cfg.APIUI.ServerURL = "https://testing.env"
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode   int
		qrWithLink bool
		response   protocol.AuthorizationRequestMessage
	}
	type testConfig struct {
		name     string
		request  AuthQRCodeRequestObject
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "should get a qr code with a link by default",
			request: AuthQRCodeRequestObject{Params: AuthQRCodeParams{Type: nil}},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
		{
			name:    "should get a qr code with a link as requested",
			request: AuthQRCodeRequestObject{Params: AuthQRCodeParams{Type: common.ToPointer(AuthQRCodeParamsTypeLink)}},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
					},
					From: issuerDID.String(),
					Typ:  "application/iden3comm-plain-json",
					Type: "https://iden3-communication.io/authorization/1.0/request",
				},
			},
		},
		{
			name:    "should get a RAW qr code as requested",
			request: AuthQRCodeRequestObject{Params: AuthQRCodeParams{Type: common.ToPointer(AuthQRCodeParamsTypeRaw)}},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: false,
				response: protocol.AuthorizationRequestMessage{
					Body: protocol.AuthorizationRequestMessageBody{
						CallbackURL: "https://testing.env/v1/authentication/callback?sessionID=",
						Reason:      "authentication",
						Scope:       make([]protocol.ZeroKnowledgeProofRequest, 0),
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
			apiURL := "/v1/authentication/qrcode"
			if tc.request.Params.Type != nil {
				apiURL += fmt.Sprintf("?type=%s", *tc.request.Params.Type)
			}
			req, err := http.NewRequest("GET", apiURL, nil)
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)
			switch tc.expected.httpCode {
			case http.StatusOK:
				var resp AuthQRCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.NotEmpty(t, resp.QrCodeLink)
				require.NotEmpty(t, resp.SessionID)

				realQR := protocol.AuthorizationRequestMessage{}
				if tc.expected.qrWithLink {
					qrLink := checkQRfetchURL(t, resp.QrCodeLink)

					// Now let's fetch the original QR using the url
					rr := httptest.NewRecorder()
					req, err := http.NewRequest(http.MethodGet, qrLink, nil)
					require.NoError(t, err)
					handler.ServeHTTP(rr, req)
					require.Equal(t, http.StatusOK, rr.Code)
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))
				} else {
					require.NoError(t, json.Unmarshal([]byte(resp.QrCodeLink), &realQR))
				}

				// Let's verify the QR body

				v := tc.expected.response

				assert.Equal(t, v.Typ, realQR.Typ)
				assert.Equal(t, v.Type, realQR.Type)
				assert.Equal(t, v.From, realQR.From)
				assert.Equal(t, v.Body.Scope, realQR.Body.Scope)
				assert.Equal(t, v.Body.Reason, realQR.Body.Reason)
				assert.True(t, strings.Contains(realQR.Body.CallbackURL, v.Body.CallbackURL))
			}
		})
	}
}

func TestServer_GetSchema(t *testing.T) {
	ctx := context.Background()
	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
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
	s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
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
				errorMsg: "Invalid format for parameter id: error unmarshaling 'someInvalidDID' text as *uuid.UUID: invalid UUID length: 14",
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
					CreatedAt: TimeUTC(s.CreatedAt),
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
				assert.InDelta(t, time.Time(tc.expected.schema.CreatedAt).UnixMilli(), time.Time(response.CreatedAt).UnixMilli(), 1000)
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

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
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
		s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
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
	s.Hash = common.CreateSchemaHash([]byte(s.URL + "#" + s.Type))
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
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
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
	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), schemaSrv, NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)

	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	server.cfg.APIUI.IssuerDID = *issuerDID
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

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
	ctx := context.Background()
	connectionsRepository := repositories.NewConnections()
	claimsRepository := repositories.NewClaims()

	connectionsService := services.NewConnection(connectionsRepository, claimsRepository, storage)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, "http://localhost", pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)

	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	issuerDID, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	server.cfg.APIUI.IssuerDID = *issuerDID
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)

	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	pubSub := pubsub.NewMock()

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.ServerUrl, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
				MtProof:        common.ToPointer(true),
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
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, "http://localhost", pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), NewConnectionsMock(), NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	handler := getHandler(context.Background(), server)

	fixture := tests.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"

	createdClaim1Proofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: true,
	}

	createdClaim2Proofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: false,
	}

	createdClaim3Proofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      false,
		Iden3SparseMerkleTreeProof: true,
	}

	createdClaim1, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, createdClaim1Proofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	createdClaim2, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, createdClaim2Proofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	createdClaim3, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, createdClaim3Proofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
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
					CreatedAt:  TimeUTC(time.Now().UTC()),
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
					CreatedAt:  TimeUTC(time.Now().UTC()),
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
					CreatedAt:  TimeUTC(time.Now().UTC()),
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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	schemaRepository := repositories.NewSchema(*storage)
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	schemaService := services.NewSchema(schemaRepository, schemaLoader)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true},
		nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	// Expires in future
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	// Expired
	claim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &past, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	// non expired, but revoked
	revoked, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, &future, typeC, nil, nil, &merklizedRootPosition,
		ports.ClaimRequestProofs{BJJSignatureProof2021: false, Iden3SparseMerkleTreeProof: true},
		nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	id, err := w3c.ParseDID(*revoked.Identifier)
	require.NoError(t, err)
	require.NoError(t, claimsService.Revoke(ctx, *id, uint64(revoked.RevNonce), "because I can"))

	handler := getHandler(ctx, server)

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
			name: "GetEthClient all implicit",
			auth: authOk,
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:   "GetEthClient all explicit",
			auth:   authOk,
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
			name:       "GetEthClient all explicit, page 1 with 2 results",
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
			name:       "GetEthClient all explicit, page 2 with 2 results",
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
			name:       "GetEthClient all explicit, page 3 with 2 results. No results",
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
			name:   "GetEthClient all from existing did",
			auth:   authOk,
			status: common.ToPointer("all"),
			did:    &claim.OtherIdentifier,
			expected: expected{
				httpCode:         http.StatusOK,
				total:            4,
				maxResults:       50,
				page:             1,
				credentialsCount: 4,
			},
		},
		{
			name:   "GetEthClient all from non existing did. Expecting empty list",
			auth:   authOk,
			status: common.ToPointer("all"),
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
			did:   &claim.OtherIdentifier,
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
			endpoint := url.URL{Path: "/v1/credentials"}
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
				queryParams = append(queryParams, "did="+*tc.did)
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
			}
		})
	}
}

func TestServer_GetCredentialQrCode(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, qrService, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), qrService, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	handler := getHandler(ctx, server)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"

	createdSIGClaimProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      true,
		Iden3SparseMerkleTreeProof: false,
	}

	createdMTPClaimProofs := ports.ClaimRequestProofs{
		BJJSignatureProof2021:      false,
		Iden3SparseMerkleTreeProof: true,
	}

	createdSIGClaim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, createdSIGClaimProofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	createdMTPClaim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, createdMTPClaimProofs, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	type expected struct {
		message    *string
		httpCode   int
		qrWithLink bool
	}

	type testConfig struct {
		name     string
		request  GetCredentialQrCodeRequestObject
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "should return an error, claim not found",
			request: GetCredentialQrCodeRequestObject{
				Id: uuid.New(),
			},
			expected: expected{
				message:  common.ToPointer("Credential not found"),
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "no mtp proof",
			request: GetCredentialQrCodeRequestObject{
				Id: createdMTPClaim.ID,
			},
			expected: expected{
				message:  common.ToPointer("State must be published before fetching MTP type credentials"),
				httpCode: http.StatusConflict,
			},
		},
		{
			name: "happy path",
			request: GetCredentialQrCodeRequestObject{
				Id: createdSIGClaim.ID,
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
			},
		},
		{
			name: "happy path with qr code of type link",
			request: GetCredentialQrCodeRequestObject{
				Id: createdSIGClaim.ID,
				Params: GetCredentialQrCodeParams{
					Type: common.ToPointer(GetCredentialQrCodeParamsTypeLink),
				},
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: true,
			},
		},
		{
			name: "happy path with qr code of type raw",
			request: GetCredentialQrCodeRequestObject{
				Id: createdSIGClaim.ID,
				Params: GetCredentialQrCodeParams{
					Type: common.ToPointer(GetCredentialQrCodeParamsTypeRaw),
				},
			},
			expected: expected{
				httpCode:   http.StatusOK,
				qrWithLink: false,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apiURL := fmt.Sprintf("/v1/credentials/%s/qrcode", tc.request.Id.String())
			if tc.request.Params.Type != nil {
				apiURL += fmt.Sprintf("?type=%s", *tc.request.Params.Type)
			}

			req, err := http.NewRequest(http.MethodGet, apiURL, nil)
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				resp := &GetCredentialQrCode200JSONResponse{}
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), resp))

				realQR := protocol.CredentialsOfferMessage{}
				if tc.expected.qrWithLink {
					qrLink := checkQRfetchURL(t, resp.QrCodeLink)

					// Now let's fetch the original QR using the url
					rr := httptest.NewRecorder()
					req, err := http.NewRequest(http.MethodGet, qrLink, nil)
					require.NoError(t, err)
					handler.ServeHTTP(rr, req)
					require.Equal(t, http.StatusOK, rr.Code)
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))
				} else {
					require.NoError(t, json.Unmarshal([]byte(resp.QrCodeLink), &realQR))
				}

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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)

	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	fixture := tests.NewFixture(storage)
	claim := fixture.NewClaim(t, did.String())
	fixture.CreateClaim(t, claim)

	usrDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	usrDID2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFBp1sRF1bFbTybVHHZQRgSWE2nKrdWeAxyZ67PdG")
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
					CreatedAt: TimeUTC(time.Now()),
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
							CreatedAt:  TimeUTC(time.Now().UTC()),
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
					CreatedAt:   TimeUTC(time.Now()),
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
				assert.InDelta(t, time.Time(tc.expected.response.CreatedAt).Unix(), time.Time(response.CreatedAt).Unix(), 10)
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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)

	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)
	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
		Hash:      common.CreateSchemaHash([]byte(schemaContext + "#" + schemaType)),
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
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject, nil, schemaType, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schemaURL, credentialSubject2, nil, schemaType, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)

	usrDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	usrDID2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFBp1sRF1bFbTybVHHZQRgSWE2nKrdWeAxyZ67PdG")
	require.NoError(t, err)

	uuid1, err := uuid.Parse("9736cf94-cd42-11ed-9618-debe37e1cbd6")
	require.NoError(t, err)

	now := time.Now()
	connID := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid1,
		IssuerDID:  *did,
		UserDID:    *usrDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  now,
		ModifiedAt: now,
	})

	uuid2, err := uuid.Parse("5736cf94-cd42-11ed-9618-debe37e1cbd6")
	require.NoError(t, err)
	connID2 := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid2,
		IssuerDID:  *did,
		UserDID:    *usrDID2,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  now.Add(1 * time.Second),
		ModifiedAt: now.Add(1 * time.Second),
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
			name: "order by wrong field",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Sort: common.ToPointer([]GetConnectionsParamsSort{"wrongField"}),
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "order by repeated fields",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Sort: common.ToPointer([]GetConnectionsParamsSort{"createdAt", "createdAt"}),
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "order by opposite fields",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Sort: common.ToPointer([]GetConnectionsParamsSort{"userID", "-userID"}),
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "order by userID desc, createdAt asc returns 2 connections",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Sort: common.ToPointer([]GetConnectionsParamsSort{"createdAt, -userID"}),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
					},
				},
			},
		},

		{
			name:    "should return 2 connections",
			auth:    authOk,
			request: GetConnectionsRequestObject{},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
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
					Items: GetConnectionsResponse{
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(time.Now()),
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
					Query: common.ToPointer("Z7gcmEoP2KppvFPCZqyzyb5tK9T6Ge"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(time.Now()),
						},
					},
				},
			},
		},
		{
			name: "should return two connections, beginning by did",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Query: common.ToPointer("did:"),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
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
					Items: GetConnectionsResponse{
						{
							Id:          connID2.String(),
							IssuerID:    did.String(),
							UserID:      usrDID2.String(),
							CreatedAt:   TimeUTC(time.Now()),
							Credentials: []Credential{},
						},
						{
							Id:          connID.String(),
							IssuerID:    did.String(),
							UserID:      usrDID.String(),
							CreatedAt:   TimeUTC(time.Now()),
							Credentials: []Credential{{}, {}},
						},
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
					Items: GetConnectionsResponse{
						{
							Id:          connID.String(),
							IssuerID:    did.String(),
							UserID:      usrDID.String(),
							CreatedAt:   TimeUTC(time.Now()),
							Credentials: []Credential{{}, {}},
						},
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
					Items: GetConnectionsResponse{
						{
							Id:          connID.String(),
							IssuerID:    did.String(),
							UserID:      usrDID.String(),
							CreatedAt:   TimeUTC(time.Now()),
							Credentials: []Credential{{}, {}},
						},
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
					Items: GetConnectionsResponse{
						{
							Id:          connID.String(),
							IssuerID:    did.String(),
							UserID:      usrDID.String(),
							CreatedAt:   TimeUTC(time.Now()),
							Credentials: []Credential{{}, {}},
						},
					},
				},
			},
		},
		{
			name: "should return no connection with not existing did and valid attributes",
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
					Items: GetConnectionsResponse{},
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
			name: "should return an error, invalid page",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Credentials: common.ToPointer(true),
					Query:       common.ToPointer("did:polygonid:polygon:mumbai:2qFVUasb8QZ1XAmD71b3NA8bzQhGs92VQEPgELYnpk"),
					Page:        common.ToPointer(uint(0)),
				},
			},
			expected: expected{
				httpCode: http.StatusBadRequest,
			},
		},
		{
			name: "should return one connection, page 1",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Page:       common.ToPointer(uint(1)),
					MaxResults: common.ToPointer(uint(1)),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
					},
				},
			},
		},
		{
			name: "should return two connection with no specified page",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					MaxResults: common.ToPointer(uint(2)),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
					},
				},
			},
		},
		{
			name: "should return one connection, page 2",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Page:       common.ToPointer(uint(2)),
					MaxResults: common.ToPointer(uint(1)),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
					},
				},
			},
		},
		{
			name: "should return 2 connections, page 1",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Page:       common.ToPointer(uint(1)),
					MaxResults: common.ToPointer(uint(30)),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
					},
				},
			},
		},
		{
			name: "should return no connections unknown page",
			auth: authOk,
			request: GetConnectionsRequestObject{
				Params: GetConnectionsParams{
					Page:       common.ToPointer(uint(10)),
					MaxResults: common.ToPointer(uint(30)),
				},
			},
			expected: expected{
				httpCode: http.StatusOK,
				response: GetConnections200JSONResponse{
					Items: GetConnectionsResponse{},
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
					Items: GetConnectionsResponse{
						{
							Id:        connID2.String(),
							IssuerID:  did.String(),
							UserID:    usrDID2.String(),
							CreatedAt: TimeUTC(now.Add(1 * time.Second)),
						},
						{
							Id:        connID.String(),
							IssuerID:  did.String(),
							UserID:    usrDID.String(),
							CreatedAt: TimeUTC(now),
						},
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
			if tc.request.Params.Page != nil {
				values.Add("page", strconv.Itoa(int(*tc.request.Params.Page)))
			}
			if tc.request.Params.MaxResults != nil {
				values.Add("max_results", strconv.Itoa(int(*tc.request.Params.MaxResults)))
			}
			if tc.request.Params.Sort != nil {
				fields := make([]string, 0)
				for _, field := range *tc.request.Params.Sort {
					fields = append(fields, string(field))
				}
				values.Add("sort", strings.Join(fields, ","))
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
				require.Equal(t, len(tc.expected.response.Items), len(response.Items))
				for i := range response.Items {
					if tc.expected.response.Items[i].Credentials != nil {
						require.NotNil(t, response.Items[i].Credentials)
						require.Equal(t, len(tc.expected.response.Items[i].Credentials), len(response.Items[i].Credentials), "connection.credentials")
					}
					assert.Equal(t, tc.expected.response.Items[i].Id, response.Items[i].Id)
					assert.Equal(t, tc.expected.response.Items[i].IssuerID, response.Items[i].IssuerID)
					assert.Equal(t, tc.expected.response.Items[i].UserID, response.Items[i].UserID)
					assert.InDelta(t, time.Time(tc.expected.response.Items[i].CreatedAt).Unix(), time.Time(response.Items[i].CreatedAt).Unix(), 10)
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
	assert.InDelta(t, time.Time(tc.CreatedAt).Unix(), time.Time(response.CreatedAt).Unix(), 10)
	if response.ExpiresAt != nil && tc.ExpiresAt != nil {
		assert.InDelta(t, time.Time(*tc.ExpiresAt).Unix(), time.Time(*response.ExpiresAt).Unix(), 10)
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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)

	fixture := tests.NewFixture(storage)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRespository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	pubSub := pubsub.NewMock()

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRespository, schemaLoader, sessionRepository, pubSub, ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
				CredentialExpiration: common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)),
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
				CredentialExpiration: common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local)),
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
				CredentialExpiration: common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local)),
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
				CredentialExpiration: common.ToPointer(time.Date(2020, 8, 15, 14, 30, 45, 100, time.Local)),
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
				CredentialExpiration: common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
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
				CredentialExpiration: common.ToPointer(time.Date(2000, 8, 15, 14, 30, 45, 100, time.Local)),
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	tomorrow := time.Now().Add(24 * time.Hour)
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, nil, true, true, CredentialSubject{"birthday": 19790911, "documentType": 12}, nil, nil)
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	require.NoError(t, err)
	hash, _ := link.Schema.Hash.MarshalText()

	linkExpired, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, common.ToPointer(tomorrow), true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
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
					Active:               link.Active,
					CredentialSubject:    CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:           common.ToPointer(TimeUTC(*link.ValidUntil)),
					Id:                   link.ID,
					IssuedClaims:         link.IssuedClaims,
					MaxIssuance:          link.MaxIssuance,
					SchemaType:           link.Schema.Type,
					SchemaUrl:            link.Schema.URL,
					Status:               LinkStatusActive,
					ProofTypes:           []string{"Iden3commRevocationStatusV1", "BJJSignature2021"},
					CreatedAt:            TimeUTC(link.CreatedAt),
					SchemaHash:           string(hash),
					CredentialExpiration: common.ToPointer(TimeUTC(tomorrow)),
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
					Active:               linkExpired.Active,
					CredentialSubject:    CredentialSubject{"birthday": 19791109, "documentType": 12, "type": schemaType, "id": "did:polygonid:polygon:mumbai:2qDDDKmo436EZGCBAvkqZjADYoNRJszkG7UymZeCHQ"},
					Expiration:           common.ToPointer(TimeUTC(*linkExpired.ValidUntil)),
					Id:                   linkExpired.ID,
					IssuedClaims:         linkExpired.IssuedClaims,
					MaxIssuance:          linkExpired.MaxIssuance,
					SchemaType:           linkExpired.Schema.Type,
					SchemaUrl:            linkExpired.Schema.URL,
					Status:               LinkStatusExceeded,
					ProofTypes:           []string{"Iden3commRevocationStatusV1", "BJJSignature2021"},
					CredentialExpiration: nil,
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
				assert.InDelta(t, time.Time(*expected.Expiration).UnixMilli(), time.Time(*response.Expiration).UnixMilli(), 1000)
				assert.Equal(t, len(expected.ProofTypes), len(response.ProofTypes))
				if expected.CredentialExpiration != nil {
					tt := time.Time(*expected.CredentialExpiration)
					tt00 := common.ToPointer(TimeUTC(time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.UTC)))
					assert.Equal(t, tt00.String(), response.CredentialExpiration.String())
				}
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
		network    = "amoy"
		BJJ        = "BJJ"
		sUrl       = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(sUrl, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	tomorrow := time.Now().Add(24 * time.Hour)
	yesterday := time.Now().Add(-24 * time.Hour)

	link1, err := linkService.Save(ctx, *did, common.ToPointer(10), &tomorrow, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12},
		&verifiable.RefreshService{
			ID:   "https://refresh.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
		&verifiable.DisplayMethod{
			ID:   "https://display.xyz",
			Type: verifiable.Iden3BasicDisplayMethodV1,
		},
	)
	require.NoError(t, err)
	linkActive := getLinkResponse(*link1)

	time.Sleep(10 * time.Millisecond)

	link2, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12},
		&verifiable.RefreshService{
			ID:   "https://revreshv2.xyz",
			Type: verifiable.Iden3RefreshService2023,
		},
		&verifiable.DisplayMethod{
			ID:   "https://display.xyz",
			Type: verifiable.Iden3BasicDisplayMethodV1,
		},
	)
	require.NoError(t, err)
	linkExpired := getLinkResponse(*link2)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	link3, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, &tomorrow, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
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
						assert.Equal(t, tc.expected.response[i].RefreshService, resp.RefreshService)
						tcCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						respCred, err := json.Marshal(tc.expected.response[i].CredentialSubject)
						require.NoError(t, err)
						assert.Equal(t, tcCred, respCred)
						assert.InDelta(t, time.Time(*tc.expected.response[i].Expiration).UnixMilli(), time.Time(*resp.Expiration).UnixMilli(), 1000)
						expectCredExpiration := common.ToPointer(TimeUTC(time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)))
						assert.Equal(t, expectCredExpiration.String(), resp.CredentialExpiration.String())
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	schemaRepository := repositories.NewSchema(*storage)
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, nil, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	iden2, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	did2, err := w3c.ParseDID(iden2.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did2
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 100, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 100, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	qrService := services.NewQrStoreService(cachex)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, qrService, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	// cfg.APIUI.ServerURL = "http://localhost/issuer-admin"

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, qrService, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	validUntil := common.ToPointer(time.Now().Add(365 * 24 * time.Hour))
	credentialExpiration := common.ToPointer(validUntil.Add(365 * 24 * time.Hour))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)

	yesterday := time.Now().Add(-24 * time.Hour)
	linkExpired, err := linkService.Save(ctx, *did, common.ToPointer(10), &yesterday, importedSchema.ID, nil, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
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
		request  CreateLinkQrCodeRequestObject
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "Wrong link id",
			request: CreateLinkQrCodeRequestObject{Id: uuid.New()},
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: link not found",
			},
		},
		{
			name: "Expired link",
			request: CreateLinkQrCodeRequestObject{
				Id: linkExpired.ID,
			},
			expected: expected{
				httpCode: http.StatusNotFound,
				message:  "error: cannot issue a credential for an expired link",
			},
		},
		{
			name: "Happy path",
			request: CreateLinkQrCodeRequestObject{
				Id: link.ID,
			},
			expected: expected{
				linkDetail: linkDetail,
				httpCode:   http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			apiURL := fmt.Sprintf("/v1/credentials/links/%s/qrcode", tc.request.Id.String())

			req, err := http.NewRequest(http.MethodPost, apiURL, tests.JSONBody(t, nil))
			require.NoError(t, err)

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				callBack := cfg.APIUI.ServerURL + "/v1/credentials/links/callback?"
				var response CreateLinkQrCode200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

				realQR := protocol.AuthorizationRequestMessage{}

				qrLink := checkQRfetchURL(t, response.QrCodeLink)

				// Now let's fetch the original QR using the url
				rr := httptest.NewRecorder()
				req, err := http.NewRequest(http.MethodGet, qrLink, nil)
				require.NoError(t, err)
				handler.ServeHTTP(rr, req)
				require.Equal(t, http.StatusOK, rr.Code)

				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))

				assert.NotNil(t, realQR.Body)
				assert.Equal(t, "authentication", realQR.Body.Reason)
				callbackArr := strings.Split(realQR.Body.CallbackURL, "sessionID")
				assert.True(t, len(callbackArr) == 2)
				assert.Equal(t, callBack, callbackArr[0])
				params := strings.Split(callbackArr[1], "linkID")
				assert.True(t, len(params) == 2)
				assert.NotNil(t, realQR.ID)
				assert.Equal(t, "https://iden3-communication.io/authorization/1.0/request", string(realQR.Type))
				assert.Equal(t, "application/iden3comm-plain-json", string(realQR.Typ))
				assert.Equal(t, cfg.APIUI.IssuerDID.String(), realQR.From)
				assert.NotNil(t, realQR.ThreadID)
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
		network    = "amoy"
		BJJ        = "BJJ"
		url        = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
		schemaType = "KYCCountryOfResidenceCredential"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	linkRepository := repositories.NewLink(*storage)
	schemaRepository := repositories.NewSchema(*storage)
	sessionRepository := repositories.NewSessionCached(cachex)

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	qrService := services.NewQrStoreService(cachex)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, qrService, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, qrService, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	linkService := services.NewLinkService(storage, claimsService, qrService, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock(), ipfsGatewayURL)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	schemaSrv := services.NewSchema(repositories.NewSchema(*storage), schemaLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer("someTitle"), uuid.NewString(), common.ToPointer("someDescription"))
	importedSchema, err := schemaSrv.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	cfg.APIUI.ServerURL = "http://localhost/issuer-admin"

	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, linkService, qrService, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	validUntil := common.ToPointer(time.Date(2023, 8, 15, 14, 30, 45, 0, time.Local))
	credentialExpiration := common.ToPointer(time.Date(2025, 8, 15, 14, 30, 45, 0, time.Local))
	link, err := linkService.Save(ctx, *did, common.ToPointer(10), validUntil, importedSchema.ID, credentialExpiration, true, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)
	handler := getHandler(ctx, server)

	sessionID := uuid.New()
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qP8KN3KRwBi37jB2ENXrWxhTo3pefaU5u5BFPbjYo")
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

	qrCodeBytes, err := json.Marshal(qrcode)
	require.NoError(t, err)

	linkDetail := getLinkResponse(*link)
	id, err := qrService.Store(ctx, qrCodeBytes, 100*time.Second)
	require.NoError(t, err)
	qrCodeLink := qrService.ToURL("http://localhost:3002", id)

	type expected struct {
		qrCode     *string
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
			state:     linkState.NewStateDone(qrCodeLink),
			expected: expected{
				linkDetail: linkDetail,
				qrCode:     common.ToPointer(qrCodeLink),
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
					assert.Equal(t, tc.expected.linkDetail.Id, response.LinkDetail.Id)
					assert.Equal(t, tc.expected.linkDetail.SchemaType, response.LinkDetail.SchemaType)
					assert.Equal(t, tc.expected.status, *response.Status)
					require.NotNil(t, response.QrCode)
					assert.Equal(t, *tc.expected.qrCode, *response.QrCode)
					qrLink := checkQRfetchURL(t, *response.QrCode)

					// Now let's fetch the original QR using the url
					rr := httptest.NewRecorder()
					req, err := http.NewRequest(http.MethodGet, qrLink, nil)
					require.NoError(t, err)
					handler.ServeHTTP(rr, req)
					require.Equal(t, http.StatusOK, rr.Code)

					// Let's verify the QR body
					realQR := protocol.CredentialsOfferMessage{}
					require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &realQR))
				}
			}
		})
	}
}

func TestServer_GetStateStatus(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"

	idenWithSignatureClaim, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	didSignatureClaim, err := w3c.ParseDID(idenWithSignatureClaim.Identifier)
	require.NoError(t, err)

	cfg1 := &config.Configuration{
		APIUI: config.APIUI{
			IssuerDID: *didSignatureClaim,
		},
	}

	serverWithSignatureClaim := NewServer(cfg1, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(didSignatureClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	handlerWithSignatureClaim := getHandler(ctx, serverWithSignatureClaim)

	idenWithMTPClaim, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	didWithMTPClaim, err := w3c.ParseDID(idenWithMTPClaim.Identifier)
	require.NoError(t, err)

	cfgWithMTPClaim := &config.Configuration{
		APIUI: config.APIUI{
			IssuerDID: *didWithMTPClaim,
		},
	}
	serverWithMTPClaim := NewServer(cfgWithMTPClaim, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(didWithMTPClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	handlerWithMTPClaim := getHandler(ctx, serverWithMTPClaim)

	idenWithRevokedClaim, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	didWithRevokedClaim, err := w3c.ParseDID(idenWithRevokedClaim.Identifier)
	require.NoError(t, err)

	cfgWithRevokedClaim := &config.Configuration{
		APIUI: config.APIUI{
			IssuerDID: *didWithRevokedClaim,
		},
	}
	serverWithRevokedClaim := NewServer(cfgWithRevokedClaim, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)
	cred, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(didWithRevokedClaim, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, true, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	require.NoError(t, err)
	require.NoError(t, claimsService.Revoke(ctx, cfgWithRevokedClaim.APIUI.IssuerDID, uint64(cred.RevNonce), "not valid"))
	handlerWithRevokedClaim := getHandler(ctx, serverWithRevokedClaim)

	type expected struct {
		response GetStateStatus200JSONResponse
		httpCode int
	}

	type testConfig struct {
		name     string
		handler  http.Handler
		auth     func() (string, string)
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:    "No auth header",
			handler: handlerWithSignatureClaim,
			auth:    authWrong,
			expected: expected{
				httpCode: http.StatusUnauthorized,
			},
		},
		{
			name:    "No states to process",
			auth:    authOk,
			handler: handlerWithSignatureClaim,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: false},
				httpCode: http.StatusOK,
			},
		},
		{
			name:    "New state to process because there is a new credential with mtp proof",
			handler: handlerWithMTPClaim,
			auth:    authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: true},
				httpCode: http.StatusOK,
			},
		},
		{
			name:    "New state to process because there is a revoked credential",
			handler: handlerWithRevokedClaim,
			auth:    authOk,
			expected: expected{
				response: GetStateStatus200JSONResponse{PendingActions: true},
				httpCode: http.StatusOK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			url := "/v1/state/status"

			req, err := http.NewRequest(http.MethodGet, url, nil)
			req.SetBasicAuth(tc.auth())
			require.NoError(t, err)

			tc.handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expected.httpCode, rr.Code)

			switch tc.expected.httpCode {
			case http.StatusOK:
				var response GetStateStatus200JSONResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
				assert.Equal(t, tc.expected.response.PendingActions, response.PendingActions)
			}
		})
	}
}

func TestServer_GetStateTransactions(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()
	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubsub.NewMock(), ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, identityService, claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

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
		network    = "amoy"
		BJJ        = "BJJ"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnections()

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
	pubSub := pubsub.NewMock()

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, schemaLoader, storage, cfg.APIUI.ServerURL, pubSub, ipfsGatewayURL, revocationStatusResolver, mediaTypeManager)
	connectionsService := services.NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	cfg.APIUI.IssuerDID = *did
	server := NewServer(&cfg, NewIdentityMock(), claimsService, NewSchemaMock(), connectionsService, NewLinkMock(), nil, NewPublisherMock(), NewPackageManagerMock(), nil, *networkResolver)

	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	merklizedRootPosition := "index"
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"

	createdCredential, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, nil, schema, credentialSubject, nil, typeC, nil, nil, &merklizedRootPosition, ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
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
