package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/http"
	networkPkg "github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/reversehash"
	"github.com/polygonid/sh-id-platform/internal/revocationstatus"
)

func TestNotification_SendNotification(t *testing.T) {
	const (
		method     = "polygonid"
		blockchain = "polygon"
		network    = "amoy"
	)
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaim()
	identityStateRepo := repositories.NewIdentityState()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	mtService := NewIdentityMerkleTrees(mtRepo)
	revocationRepository := repositories.NewRevocation()
	connectionsRepository := repositories.NewConnection()

	reader := common.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	rhsFactory := reversehash.NewFactory(*networkResolver, reversehash.DefaultRHSTimeOut)
	revocationStatusResolver := revocationstatus.NewRevocationStatusResolver(*networkResolver)
	identityService := NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)

	mediaTypeManager := NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	credentialsService := NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, docLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGateway, revocationStatusResolver, mediaTypeManager, cfg.UniversalLinks)
	connectionsService := NewConnection(connectionsRepository, claimsRepo, storage)
	iden, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	did, err := w3c.ParseDID(iden.Identifier)
	require.NoError(t, err)

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	notificationGateway := gateways.NewPushNotificationClient(http.DefaultHTTPClientWithRetry)
	notificationService := NewNotification(notificationGateway, connectionsService, credentialsService)

	fixture := repositories.NewFixture(storage)
	credID := fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(did.String()),
		Issuer:          did.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246564",
	})

	t.Run("should get an error, non existing credential", func(t *testing.T) {
		ev := event.CreateCredential{CredentialIDs: []string{uuid.NewString()}, IssuerID: did.String()}
		message, err := ev.Marshal()
		require.NoError(t, err)
		assert.Error(t, notificationService.SendCreateCredentialNotification(ctx, message))
	})

	t.Run("should get an error, existing credential but not existing connection", func(t *testing.T) {
		ev := event.CreateCredential{CredentialIDs: []string{credID.String()}, IssuerID: did.String()}
		message, err := ev.Marshal()
		require.NoError(t, err)
		assert.Error(t, notificationService.SendCreateCredentialNotification(ctx, message))
	})

	t.Run("should get an error,wrong credential id", func(t *testing.T) {
		ev := event.CreateCredential{CredentialIDs: []string{"wrong id"}, IssuerID: did.String()}
		message, err := ev.Marshal()
		require.NoError(t, err)
		assert.Error(t, notificationService.SendCreateCredentialNotification(ctx, message))
	})
}
