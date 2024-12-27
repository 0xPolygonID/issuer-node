package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	networkPkg "github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/pubsub"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/reversehash"
	"github.com/polygonid/sh-id-platform/internal/revocationstatus"
)

func TestSchema_ImportSchema(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const title = "someTitle"
	const description = "someDescription"
	const urlLD = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
	const schemaType = "KYCCountryOfResidenceCredential"
	const did = "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"

	version := uuid.NewString()
	ctx := context.Background()
	repo := repositories.NewSchema(*storage)

	issuerDID, err := w3c.ParseDID(did)
	require.NoError(t, err)

	displayMethodRepository := repositories.NewDisplayMethod(*storage)
	displayMethodService := NewDisplayMethod(displayMethodRepository)

	expectHash := utils.CreateSchemaHash([]byte(urlLD + "#" + schemaType))

	s := NewSchema(repo, docLoader, displayMethodService)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer(title), version, common.ToPointer(description), nil)
	got, err := s.ImportSchema(ctx, *issuerDID, iReq)
	require.NoError(t, err)
	_, err = uuid.Parse(got.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, url, got.URL)
	assert.Equal(t, schemaType, got.Type)
	assert.Equal(t, did, got.IssuerDID.String())
	assert.Equal(t, expectHash, got.Hash)
	assert.Len(t, got.Words, 3)
	assert.InDelta(t, time.Now().UnixMilli(), got.CreatedAt.UnixMilli(), 10)
	assert.Equal(t, title, *got.Title)
	assert.Equal(t, description, *got.Description)
	assert.Equal(t, version, got.Version)

	updatedSchema, err := s.GetByID(ctx, *issuerDID, got.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSchema.Words, 4)
}

func TestSchema_Update(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const title = "someTitle"
	const description = "someDescription"
	const urlLD = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
	const schemaType = "KYCCountryOfResidenceCredential"

	version := uuid.NewString()
	ctx := context.Background()
	identityRepo := repositories.NewIdentity()
	claimsRepo := repositories.NewClaim()
	mtRepo := repositories.NewIdentityMerkleTreeRepository()
	identityStateRepo := repositories.NewIdentityState()
	revocationRepository := repositories.NewRevocation()
	schemaRepository := repositories.NewSchema(*storage)
	mtService := NewIdentityMerkleTrees(mtRepo)
	connectionsRepository := repositories.NewConnection()
	keyRepository := repositories.NewKey(*storage)
	displayMethodRepository := repositories.NewDisplayMethod(*storage)
	displayMethodService := NewDisplayMethod(displayMethodRepository)

	reader := common.CreateFile(t)
	networkResolver, err := networkPkg.NewResolver(ctx, cfg, keyStore, reader)
	require.NoError(t, err)

	rhsFactory := reversehash.NewFactory(*networkResolver, reversehash.DefaultRHSTimeOut)
	revocationStatusResolver := revocationstatus.NewRevocationStatusResolver(*networkResolver)
	identityService := NewIdentity(keyStore, identityRepo, mtRepo, identityStateRepo, mtService, nil, claimsRepo, revocationRepository, connectionsRepository, storage, nil, nil, pubsub.NewMock(), *networkResolver, rhsFactory, revocationStatusResolver, keyRepository)
	schemaService := NewSchema(schemaRepository, docLoader, displayMethodService)

	identity, err := identityService.Create(ctx, "polygon-test", &ports.DIDCreationOptions{Method: method, Blockchain: blockchain, Network: net, KeyType: BJJ})
	assert.NoError(t, err)
	issuerDID, err := w3c.ParseDID(identity.Identifier)
	require.NoError(t, err)

	displayMethodID, err := displayMethodService.Save(ctx, *issuerDID, "display method", "url", nil)
	require.NoError(t, err)

	displayMethodID2, err := displayMethodService.Save(ctx, *issuerDID, "display method 2", "url", nil)
	require.NoError(t, err)

	expectHash := utils.CreateSchemaHash([]byte(urlLD + "#" + schemaType))
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer(title), version, common.ToPointer(description), displayMethodID)

	got, err := schemaService.ImportSchema(ctx, *issuerDID, iReq)
	require.NoError(t, err)
	_, err = uuid.Parse(got.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, url, got.URL)
	assert.Equal(t, schemaType, got.Type)
	assert.Equal(t, issuerDID.String(), got.IssuerDID.String())
	assert.Equal(t, expectHash, got.Hash)
	assert.Len(t, got.Words, 3)
	assert.InDelta(t, time.Now().UnixMilli(), got.CreatedAt.UnixMilli(), 10)
	assert.Equal(t, title, *got.Title)
	assert.Equal(t, description, *got.Description)
	assert.Equal(t, version, got.Version)
	assert.Equal(t, displayMethodID, got.DisplayMethodID)

	got.DisplayMethodID = displayMethodID2
	require.NoError(t, schemaService.Update(ctx, got))

	updatedSchema, err := schemaService.GetByID(ctx, *issuerDID, got.ID)
	require.NoError(t, err)
	_, err = uuid.Parse(updatedSchema.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, url, updatedSchema.URL)
	assert.Equal(t, schemaType, updatedSchema.Type)
	assert.Equal(t, issuerDID.String(), updatedSchema.IssuerDID.String())
	assert.Equal(t, expectHash, updatedSchema.Hash)
	assert.Len(t, updatedSchema.Words, 4)
	assert.InDelta(t, time.Now().UnixMilli(), updatedSchema.CreatedAt.UnixMilli(), 15)
	assert.Equal(t, title, *updatedSchema.Title)
	assert.Equal(t, description, *updatedSchema.Description)
	assert.Equal(t, version, updatedSchema.Version)
	assert.Equal(t, displayMethodID2, updatedSchema.DisplayMethodID)
}
