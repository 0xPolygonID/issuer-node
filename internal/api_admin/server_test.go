package api_admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/session"
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

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaService, NewPublisherMock(), NewPackageManagerMock(), &health.Status{})
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
	server := NewServer(&cfg, NewIdentityMock(), NewClaimsMock(), NewSchemaMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	handler := getHandler(context.Background(), server)

	type expected struct {
		httpCode int
		message  *string
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
				message:  common.ToPointer("Cannot proceed with empty sessionID"),
			},
		},
		{
			name:      "should get an error no body",
			sessionID: common.ToPointer("12345"),
			expected: expected{
				httpCode: http.StatusBadRequest,
				message:  common.ToPointer("Cannot proceed with empty body"),
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
				assert.Equal(t, *tc.expected.message, response.Message)
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
	sessionManager := session.Cached(cachex)

	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, sessionManager)
	server := NewServer(&cfg, identityService, NewClaimsMock(), NewSchemaMock(), NewPublisherMock(), NewPackageManagerMock(), nil)
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)
	server.cfg.APIUI.IssuerDID = issuerDID
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
						CallbackUrl: "https://testing.env//v1/authentication/callback?sessionID=",
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
