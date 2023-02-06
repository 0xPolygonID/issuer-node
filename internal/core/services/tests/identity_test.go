package services_tests

import (
	"context"
	"testing"

	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	method     = "polygonid"
	blockchain = "polygon"
	network    = "mumbai"
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
	identityService := services.NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, claimsRepo, revocationRepository, storage, rhsp)
	schemaService := services.NewSchema(loader.OnceFactory(loader.HTTPFactory))

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

	identity, err := identityService.Create(ctx, method, blockchain, network, "http://localhost:3001")
	assert.NoError(t, err)

	identity2, err := identityService.Create(ctx, method, blockchain, network, "http://localhost:3001")
	assert.NoError(t, err)

	schema := "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	did, err := core.ParseDID(identity.Identifier)
	assert.NoError(t, err)
	did2, err := core.ParseDID(identity2.Identifier)
	assert.NoError(t, err)
	did3, err := core.ParseDID("did:polygonid:polygon:mumbai:2qD6cqGpLX2dibdFuKfrPxGiybi3wKa8RbR4onw49H")
	assert.NoError(t, err)
	credentialSubject := map[string]any{
		"id":           "did:polygonid:polygon:mumbai:2qE1BZ7gcmEoP2KppvFPCZqyzyb5tK9T6Gec5HFANQ",
		"birthday":     19960424,
		"documentType": 2,
	}
	typeC := "KYCAgeCredential"
	expiration := int64(12345)

	merklizedRootPosition := "index"
	_, err = claimsService.CreateClaim(context.Background(), ports.NewCreateClaimRequest(did, schema, credentialSubject, &expiration, typeC, nil, nil, &merklizedRootPosition))
	assert.NoError(t, err)

	type testConfig struct {
		name            string
		did             *core.DID
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
			identityState, err := identityService.UpdateState(ctx, tc.did)
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
