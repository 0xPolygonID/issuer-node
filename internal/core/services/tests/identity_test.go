package services_tests

import (
	"context"
	"testing"
	"time"

	commonEth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	method     = "polygonid"
	blockchain = "polygon"
	network    = "mumbai"
	BJJ        = "BJJ"
	host       = "https://host.com"
)

func Test_identity_UpdateState(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	connectionsRepository := repositories.NewConnections()
	rhsFactory := reverse_hash.NewFactory(cfg.CredentialStatus.RHS.URL, nil, commonEth.HexToAddress(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract), reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(cfg.CredentialStatus)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), cfg.CredentialStatus, rhsFactory, revocationStatusResolver)
	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, docLoader, storage, cfg.CredentialStatus.Iden3CommAgentStatus.GetURL(), pubsub.NewMock(), ipfsGateway, revocationStatusResolver)

	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := w3c.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	t.Run("should update state", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(true), common.ToPointer(true), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		previousStateIdentity, _ := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		identityState, err := identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
		assert.Equal(t, did.String(), identityState.Identifier)
		assert.NotNil(t, identityState.State)
		assert.Equal(t, domain.StatusCreated, identityState.Status)
		assert.NotNil(t, identityState.StateID)
		assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
		assert.NotNil(t, identityState.RootOfRoots)
		assert.NotNil(t, identityState.ClaimsTreeRoot)
		assert.NotNil(t, identityState.RevocationTreeRoot)
	})

	t.Run("should update state for a new credential with mtp", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(false), common.ToPointer(true), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		previousStateIdentity, _ := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		identityState, err := identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
		assert.Equal(t, did.String(), identityState.Identifier)
		assert.NotNil(t, identityState.State)
		assert.Equal(t, domain.StatusCreated, identityState.Status)
		assert.NotNil(t, identityState.StateID)
		assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
		assert.NotNil(t, identityState.RootOfRoots)
		assert.NotNil(t, identityState.ClaimsTreeRoot)
		assert.NotNil(t, identityState.RevocationTreeRoot)
	})

	t.Run("should return success after revoke a MTP credential", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		claim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(false), common.ToPointer(true), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		previousStateIdentity, _ := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		identityState, err := identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
		assert.Equal(t, did.String(), identityState.Identifier)
		assert.NotNil(t, identityState.State)
		assert.Equal(t, domain.StatusCreated, identityState.Status)
		assert.NotNil(t, identityState.StateID)
		assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
		assert.NotNil(t, identityState.RootOfRoots)
		assert.NotNil(t, identityState.ClaimsTreeRoot)
		assert.NotNil(t, identityState.RevocationTreeRoot)

		assert.NoError(t, claimsService.Revoke(ctx, *did, uint64(claim.RevNonce), ""))
		_, err = identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
	})

	t.Run("should return pass after creating two credentials", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		claimMTP, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(false), common.ToPointer(true), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		previousStateIdentity, _ := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		identityState, err := identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
		assert.Equal(t, did.String(), identityState.Identifier)
		assert.NotNil(t, identityState.State)
		assert.Equal(t, domain.StatusCreated, identityState.Status)
		assert.NotNil(t, identityState.StateID)
		assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
		assert.NotNil(t, identityState.RootOfRoots)
		assert.NotNil(t, identityState.ClaimsTreeRoot)
		assert.NotNil(t, identityState.RevocationTreeRoot)

		assert.NoError(t, claimsService.Revoke(ctx, *did, uint64(claimMTP.RevNonce), ""))
		_, err = identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)

		claimSIG, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(true), common.ToPointer(false), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		_, err = identityService.UpdateState(ctx, *did)
		assert.Error(t, err)

		assert.NoError(t, claimsService.Revoke(ctx, *did, uint64(claimSIG.RevNonce), ""))
		identityState, err = identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
		previousStateIdentity, err = identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		assert.NoError(t, err)
		assert.Equal(t, did.String(), identityState.Identifier)
		assert.NotNil(t, identityState.State)
		assert.Equal(t, domain.StatusCreated, identityState.Status)
		assert.NotNil(t, identityState.StateID)
		assert.Equal(t, previousStateIdentity.State, identityState.PreviousState)
		assert.NotNil(t, identityState.RootOfRoots)
		assert.NotNil(t, identityState.ClaimsTreeRoot)
		assert.NotNil(t, identityState.RevocationTreeRoot)
	})

	t.Run("should get an error creating credential with sig proof", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(true), common.ToPointer(false), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		_, err = identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		assert.NoError(t, err)
		_, err = identityService.UpdateState(ctx, *did)
		assert.Error(t, err)
	})

	t.Run("should update state after revoke credential with sig proof", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		claim, err := claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			common.ToPointer(true), common.ToPointer(false), nil, false,
			verifiable.SparseMerkleTreeProof, nil, nil, nil))

		assert.NoError(t, err)
		_, err = identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
		assert.NoError(t, err)
		_, err = identityService.UpdateState(ctx, *did)
		assert.Error(t, err)

		assert.NoError(t, claimsService.Revoke(ctx, *did, uint64(claim.RevNonce), ""))
		_, err = identityService.UpdateState(ctx, *did)
		assert.NoError(t, err)
	})
}

func Test_identity_GetByDID(t *testing.T) {
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)
	connectionsRepository := repositories.NewConnections()
	rhsFactory := reverse_hash.NewFactory(cfg.CredentialStatus.RHS.URL, nil, commonEth.HexToAddress(cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract), reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(cfg.CredentialStatus)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), cfg.CredentialStatus, rhsFactory, revocationStatusResolver)
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
