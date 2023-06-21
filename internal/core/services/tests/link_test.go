package services_tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/repositories"
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
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock())
	schemaLoader := loader.CachedFactory(loader.MultiProtocolFactory("https://gateway.ipfs.io"), cachex)
	sessionRepository := repositories.NewSessionCached(cachex)
	schemaService := services.NewSchema(schemaRepository, schemaLoader)
	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "https://host.com",
	}
	claimsService := services.NewClaim(
		claimsRepo,
		identityService,
		mtService,
		identityStateRepo,
		schemaLoader,
		storage,
		claimsConf,
		pubsub.NewMock())

	identity, err := identityService.Create(ctx, method, blockchain, network, "http://localhost:3001")
	assert.NoError(t, err)

	identity2, err := identityService.Create(ctx, method, blockchain, network, "http://localhost:3001")
	assert.NoError(t, err)

	schemaUrl := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := core.ParseDID(identity.Identifier)
	assert.NoError(t, err)

	schema, err := schemaService.ImportSchema(ctx, *did, schemaUrl, "KYCAgeCredential")
	assert.NoError(t, err)
	did2, err := core.ParseDID(identity2.Identifier)
	assert.NoError(t, err)
	//
	//did3, err := core.ParseDID("did:polygonid:polygon:mumbai:2qD6cqGpLX2dibdFuKfrPxGiybi3wKa8RbR4onw49H")
	//assert.NoError(t, err)

	userDID1 := core.DID{}
	assert.NoError(t, userDID1.SetString("did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ"))
	credentialSubject := map[string]any{
		"id":           userDID1.String(),
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "index"
	_, err = claimsService.Save(context.Background(), ports.NewCreateClaimRequest(did, schemaUrl, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	assert.NoError(t, err)

	linkRepository := repositories.NewLink(*storage)
	linkService := services.NewLinkService(storage, claimsService, claimsRepo, linkRepository, schemaRepository, schemaLoader, sessionRepository, pubsub.NewMock())

	tomorrow := time.Now().Add(24 * time.Hour)
	nextWeek := time.Now().Add(7 * 24 * time.Hour)

	link, err := linkService.Save(ctx, *did, common.ToPointer(100), &tomorrow, schema.ID, &nextWeek, true, false, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)

	link2, err := linkService.Save(ctx, *did, common.ToPointer(100), &tomorrow, schema.ID, &nextWeek, false, true, domain.CredentialSubject{"birthday": 19791109, "documentType": 12})
	assert.NoError(t, err)

	type expected struct {
		err          error
		status       string
		issuedClaims int
	}

	type testConfig struct {
		name     string
		did      core.DID
		userDID  core.DID
		LinkID   uuid.UUID
		expected expected
	}

	for _, tc := range []testConfig{
		{
			name:    "should return status done",
			did:     *did,
			userDID: userDID1,
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
			userDID: userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err:          nil,
				status:       "pendingPublish",
				issuedClaims: 1,
			},
		},
		{
			name:    "should return error",
			did:     *did,
			userDID: userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err:          services.ErrClaimAlreadyIssued,
				status:       "",
				issuedClaims: 1,
			},
		},
		{
			name:    "should return error wrong did",
			did:     *did2,
			userDID: userDID1,
			LinkID:  link2.ID,
			expected: expected{
				err: errors.New("link does not exist"),
			},
		},
		{
			name:    "should return error wrong link id",
			did:     *did,
			userDID: userDID1,
			LinkID:  uuid.New(),
			expected: expected{
				err: errors.New("link does not exist"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sessionID := uuid.New().String()
			err := linkService.IssueClaim(ctx, sessionID, tc.did, tc.userDID, tc.LinkID, "host_url")
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
