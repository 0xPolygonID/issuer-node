package services_tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/issuer-node/internal/common"
	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/core/services"
	"github.com/polygonid/issuer-node/internal/repositories"
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
	repo := repositories.NewSchemaInMemory()

	issuerDID, err := w3c.ParseDID(did)
	require.NoError(t, err)

	expectHash := utils.CreateSchemaHash([]byte(urlLD + "#" + schemaType))

	s := services.NewSchema(repo, docLoader)
	iReq := ports.NewImportSchemaRequest(url, schemaType, common.ToPointer(title), version, common.ToPointer(description))
	got, err := s.ImportSchema(ctx, *issuerDID, iReq)
	require.NoError(t, err)
	_, err = uuid.Parse(got.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, url, got.URL)
	assert.Equal(t, schemaType, got.Type)
	assert.Equal(t, did, got.IssuerDID.String())
	assert.Equal(t, expectHash, got.Hash)
	assert.Len(t, got.Words, 3)
	assert.InDelta(t, time.Now().UnixMilli(), got.CreatedAt.UnixMilli(), 1)
	assert.Equal(t, title, *got.Title)
	assert.Equal(t, description, *got.Description)
	assert.Equal(t, version, got.Version)
}
