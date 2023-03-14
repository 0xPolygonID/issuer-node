package services_tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/utils"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaAdmin_ImportSchema(t *testing.T) {
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	const urlLD = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld"
	const schemaType = "KYCCountryOfResidenceCredential"
	const did = "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"
	ctx := context.Background()
	repo := repositories.NewSchemaInMemory()

	issuerDID := core.DID{}
	require.NoError(t, issuerDID.SetString(did))

	expectHash := utils.CreateSchemaHash([]byte(urlLD + "#" + schemaType))

	s := services.NewSchemaAdmin(repo, loader.HTTPFactory)
	got, err := s.ImportSchema(ctx, issuerDID, url, schemaType)
	require.NoError(t, err)
	_, err = uuid.Parse(got.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, url, got.URL)
	assert.Equal(t, schemaType, got.Type)
	assert.Equal(t, did, got.IssuerDID.String())
	assert.Equal(t, expectHash, got.Hash)
	assert.Len(t, got.Attributes, 3)
	assert.InDelta(t, time.Now().UnixMilli(), got.CreatedAt.UnixMilli(), 1)
}
