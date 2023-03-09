package api_admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/health"
	"github.com/polygonid/sh-id-platform/internal/loader"
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
	identityService := services.NewIdentity(&KMSMock{}, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(loader.CachedFactory(loader.HTTPFactory, cachex))
	schemaAdminService := services.NewSchemaAdmin(repositories.NewSchema(storage.Pgx))

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "host",
	}
	claimsService := services.NewClaim(claimsRepo, schemaService, identityService, mtService, identityStateRepo, storage, claimsConf)

	server := NewServer(&cfg, identityService, claimsService, schemaAdminService, NewPublisherMock(), NewPackageManagerMock(), &health.Status{})
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
