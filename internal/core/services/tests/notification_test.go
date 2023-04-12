package services_tests

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/http"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func TestNotification_SendNotification(t *testing.T) {
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
	credentialsService := services.NewClaim(claimsRepo, identityService, mtService, identityStateRepo, schemaLoader, storage, claimsConf)
	connectionsService := services.NewConnection(connectionsRepository, storage)
	iden, err := identityService.Create(ctx, method, blockchain, network, "polygon-test")
	require.NoError(t, err)

	did, err := core.ParseDID(iden.Identifier)
	require.NoError(t, err)

	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	notificationGateway := gateways.NewPushNotificationClient(http.DefaultHTTPClientWithRetry)
	notificationService := services.NewNotification(notificationGateway, connectionsService, credentialsService)

	fixture := tests.NewFixture(storage)
	credID := fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(did.String()),
		Issuer:          did.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246564",
	})

	t.Run("should get an error, non existing credential", func(t *testing.T) {
		assert.Error(t, notificationService.SendCreateCredentialNotification(ctx, pubsub.CreateCredentialEvent{CredentialID: uuid.NewString(), IssuerID: did.String()}))
	})

	t.Run("should get an error, existing credential but not existing connection", func(t *testing.T) {
		assert.Error(t, notificationService.SendCreateCredentialNotification(ctx, pubsub.CreateCredentialEvent{CredentialID: credID.String(), IssuerID: did.String()}))
	})
}
