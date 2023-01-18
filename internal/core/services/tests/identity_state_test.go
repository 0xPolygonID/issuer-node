package services_tests

import (
	"context"
	"testing"

	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func Test_identityState_UpdateIdentityClaims(t *testing.T) {
	// given
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaims()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	mtService := services.NewIdentityMerkleTrees(mtRepo)

	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, storage)
	schemaService := services.NewSchema(storage)

	claimsConf := services.ClaimCfg{
		RHSEnabled: false,
		Host:       "https://host.com",
	}
	claimsService := services.NewClaim(
		claimsRepo,
		schemaService,
		identityService,
		mtService,
		identityStateRepo,
		storage,
		claimsConf,
	)

	identityStateService := services.NewIdentityState(identityStateRepo, mtService, claimsRepo, revocationRepository, storage)

	identity, err := identityService.Create(ctx, "http://localhost:3001")
	assert.NoError(t, err)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, _ := core.ParseDID(identity.Identifier)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	expiration := int64(12345)

	merklizedRootPosition := "value"
	_, err = claimsService.CreateClaim(context.Background(), ports.NewCreateClaimRequest(did, schema, credentialSubject, &expiration, typeC, nil, nil, &merklizedRootPosition))
	assert.NoError(t, err)

	previousState, err := identityStateRepo.GetLatestStateByIdentifier(ctx, storage.Pgx, did)
	assert.NoError(t, err)

	// when
	identityState, err := identityStateService.UpdateIdentityClaims(ctx, did)

	// then
	assert.NoError(t, err)
	assert.Equal(t, identity.Identifier, identityState.Identifier)
	assert.NotNil(t, identityState.State)
	assert.NotNil(t, identityState.StateID)
	assert.Equal(t, previousState.State, identityState.PreviousState)
	assert.NotNil(t, identityState.RootOfRoots)
	assert.NotNil(t, identityState.ClaimsTreeRoot)
	assert.NotNil(t, identityState.RevocationTreeRoot)
	assert.Equal(t, domain.StatusCreated, identityState.Status)
}
