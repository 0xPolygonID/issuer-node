package services_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	commonEth "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

func Test_link_issueClaim(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	schemaRepository := repositories.NewSchema(*storage)
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	connectionsRepository := repositories.NewConnections()
	rhsFactory := reverse_hash.NewFactory(cfg.CredentialStatus.RHS.URL, nil, commonEth.HexToAddress(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract), reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(cfg.CredentialStatus)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), cfg.CredentialStatus, rhsFactory, revocationStatusResolver, *cfg.AutoPublishingToOnChainRHS)
	sessionRepository := repositories.NewSessionCached(cachex)
	schemaService := services.NewSchema(schemaRepository, docLoader)

	mediaTypeManager := services.NewMediaTypeManager(
		map[iden3comm.ProtocolMessage][]string{
			protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
			protocol.RevocationStatusRequestMessageType: {"*"},
		},
		true,
	)

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, docLoader, storage, cfg.CredentialStatus.Iden3CommAgentStatus.GetURL(), pubsub.NewMock(), ipfsGateway, revocationStatusResolver, mediaTypeManager)
	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)

	identity2, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)

	schemaUrl := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := w3c.ParseDID(identity.Identifier)
	require.NoError(t, err)

	iReq := ports.NewImportSchemaRequest(schemaUrl, "KYCAgeCredential", common.ToPointer("some title"), uuid.NewString(), common.ToPointer("some description"))
	schema, err := schemaService.ImportSchema(ctx, *did, iReq)
	assert.NoError(t, err)
	did2, err := w3c.ParseDID(identity2.Identifier)
	require.NoError(t, err)

	userDID1, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ")
	require.NoError(t, err)

	credentialSubject := map[string]any{
		"id":           userDID1.String(),
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "index"
	_, err = claimsService.Save(context.Background(), ports.NewCreateClaimRequest(did, schemaUrl, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false, verifiable.Iden3commRevocationStatusV1, nil, nil, nil))
	assert.NoError(t, err)

	linkRepository := repositories.NewLink(*storage)
	qrService := services.NewQrStoreService(cachex)
	linkService := services.NewLinkService(storage, claimsService, qrService, claimsRepo, linkRepository, schemaRepository, docLoader, sessionRepository, pubsub.NewMock(), ipfsGateway)

	tomorrow := time.Now().Add(24 * time.Hour)
	nextWeek := time.Now().Add(7 * 24 * time.Hour)

	link, err := linkService.Save(ctx, *did, common.ToPointer(100), &tomorrow, schema.ID, &nextWeek, true, false, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)

	link2, err := linkService.Save(ctx, *did, common.ToPointer(100), &tomorrow, schema.ID, &nextWeek, false, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12}, nil, nil)
	assert.NoError(t, err)

	type expected struct {
		err          error
		status       string
		issuedClaims int
	}

	type testConfig struct {
		name     string
		did      w3c.DID
		userDID  w3c.DID
		LinkID   uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "should return status done",
			did:     *did,
			userDID: *userDID1,
			LinkID:  link.ID,
			expected: expected{
				err:          nil,
				status:       "done",
				issuedClaims: 1,
			},
		},
		{
			name:    "should return status pending to publish",
			did:     *did,
			userDID: *userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err:          nil,
				status:       "pendingPublish",
				issuedClaims: 1,
			},
		},
		{
			name:    "should return status pending to publish for same link",
			did:     *did,
			userDID: *userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err:          nil,
				status:       "pendingPublish",
				issuedClaims: 1,
			},
		},
		{
			name:    "should return error wrong did",
			did:     *did2,
			userDID: *userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err: errors.New("link does not exist"),
			},
		},
		{
			name:    "should return error wrong link id",
			did:     *did,
			userDID: *userDID1,
			LinkID:  uuid.New(),
			expected: expected{
				err: errors.New("link does not exist"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sessionID := uuid.New().String()
			err := linkService.IssueClaim(ctx, sessionID, tc.did, tc.userDID, tc.LinkID, "host_url", verifiable.Iden3commRevocationStatusV1)
			if tc.expected.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expected.err, err)
			} else {
				status, err := sessionRepository.GetLink(ctx, linkState.CredentialStateCacheKey(tc.LinkID.String(), sessionID))
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.status, status.Status)
				claims, err := claimsRepo.GetClaimsIssuedForUser(ctx, storage.Pgx, tc.did, tc.userDID, tc.LinkID)
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.issuedClaims, len(claims))
			}
		})
	}
}
