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
