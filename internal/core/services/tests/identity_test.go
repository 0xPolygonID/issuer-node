package services_tests

import (
	"context"
	"testing"
	"time"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	method     = "polygonid"
	blockchain = "polygon"
	network    = "mumbai"
	BJJ        = "BJJ"
)

func Test_identity_UpdateState(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	claimsConf := services.CredentialRevocationSettings{
		RHSEnabled: false,
		Host:       "https://host.com",
	}
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock(), claimsConf)
	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, docLoader, storage, claimsConf, pubsub.NewMock(), ipfsGateway)

	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	identity2, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := w3c.ParseDID(identity.Identifier)
	assert.NoError(t, err)
	did2, err := w3c.ParseDID(identity2.Identifier)
	assert.NoError(t, err)
	did3, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qD6cqGpLX2dibdFuKfrPxGiybi3wKa8RbR4onw49H")
	assert.NoError(t, err)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	merklizedRootPosition := "index"
	_, err = claimsService.Save(context.Background(), ports.NewCreateClaimRequest(did, schema, credentialSubject, common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition, common.ToPointer(true), common.ToPointer(true), nil, false))
	assert.NoError(t, err)

	type testConfig struct {
		name            string
		did             *w3c.DID
		shouldReturnErr bool
	}

	for _, tc := range []testConfig{
		{
			name:            "should get a new state for identity with a claim",
			did:             did,
			shouldReturnErr: false,
		},
		{
			name:            "should get a new state for identity without claim",
			did:             did2,
			shouldReturnErr: false,
		},
		{
			name:            "should return an error",
			did:             did3,
			shouldReturnErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previousStateIdentity, _ := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, tc.did)
			identityState, err := identityService.UpdateState(ctx, *tc.did)
			if tc.shouldReturnErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, err)
				assert.Equal(t, tc.did.String(), identityState.Identifier)
				assert.NotNil(t, identityState.State)
				assert.Equal(t, domain.StatusCreated, identityState.Status)
				assert.NotNil(t, identityState.StateID)
				assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
				assert.NotNil(t, identityState.RootOfRoots)
				assert.NotNil(t, identityState.ClaimsTreeRoot)
				assert.NotNil(t, identityState.RevocationTreeRoot)
				assert.Equal(t, domain.StatusCreated, identityState.Status)
			}
		})
	}
}

func Test_identity_GetByDID(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	rhsp := reverse_hash.NewRhsPublisher(nil, false)
	connectionsRepository := repositories.NewConnections()
	claimsConf := services.CredentialRevocationSettings{
		RHSEnabled: false,
		Host:       "https://host.com",
	}
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, rhsp, nil, nil, pubsub.NewMock(), claimsConf)
	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	assert.NoError(t, err)

	did, err := w3c.ParseDID(identity.Identifier)
	assert.NoError(t, err)

	did2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qD6cqGpLX2dibdFuKfrPxGiybi3wKa8RbR4onw49H")
	assert.NoError(t, err)

	type testConfig struct {
		name            string
		did             *w3c.DID
		shouldReturnErr bool
	}

	for _, tc := range []testConfig{
		{
			name:            "should get the identity",
			did:             did,
			shouldReturnErr: false,
		},
		{
			name:            "should return an error",
			did:             did2,
			shouldReturnErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			identityState, err := identityService.GetByDID(ctx, *tc.did)
			if tc.shouldReturnErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NoError(t, err)
				assert.Equal(t, tc.did.String(), identityState.Identifier)
			}
		})
	}
}
