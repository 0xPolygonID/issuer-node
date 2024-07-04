package services_tests

import (
	"context"
	"testing"
	"time"

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
	"github.com/polygonid/sh-id-platform/pkg/helpers"
	networkPkg "github.com/polygonid/sh-id-platform/pkg/network"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	method     = "polygonid"
	blockchain = "polygon"
	network    = "amoy"
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

	claimsService := services.NewClaim(claimsRepo, identityService, nil, mtService, identityStateRepo, docLoader, storage, cfg.ServerUrl, pubsub.NewMock(), ipfsGateway, revocationStatusResolver, mediaTypeManager)

	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: network, KeyType: BJJ})
	require.NoError(t, err)
	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := w3c.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:amoy:2qSuD8ZDpsAG3s8WJjwzqhMsqGLz8RUG1BHVUe3Gwu",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"

	t.Run("should update state", func(t *testing.T) {
		ctx := context.Background()
		merklizedRootPosition := "index"
		_, err = claimsService.Save(ctx, ports.NewCreateClaimRequest(did, schema, credentialSubject,
			common.ToPointer(time.Now()), typeC, nil, nil, &merklizedRootPosition,
			ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: true}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: false, Iden3SparseMerkleTreeProof: true}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: false, Iden3SparseMerkleTreeProof: true}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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
			ports.ClaimRequestProofs{BJJSignatureProof2021: true, Iden3SparseMerkleTreeProof: false}, nil, false,
			verifiable.Iden3commRevocationStatusV1, nil, nil, nil))

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

	reader := helpers.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	rhsFactory := reverse_hash.NewFactory(*networkResolver, reverse_hash.DefaultRHSTimeOut)
	revocationStatusResolver := revocation_status.NewRevocationStatusResolver(*networkResolver)
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver)
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
