package tests

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestGetSchema(t *testing.T) {
	rand.NewSource(time.Now().Unix())
	ctx := context.Background()
	store := repositories.NewSchema(storage.Pgx)
	did := core.DID{}
	// Create a schemaHash
	i := &big.Int{}
	i.SetInt64(rand.Int63())

	require.NoError(t, did.SetString("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"))
	schema1 := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  did,
		URL:        "https://an.url.org/index.html",
		Type:       "schemaType",
		Hash:       core.NewSchemaHashFromInt(i),
		Attributes: domain.SchemaAttrs{"field1", "field2", "fieldn"},
		CreatedAt:  time.Now(),
	}
	require.NoError(t, store.Save(ctx, schema1))

	schema2, err := store.GetById(ctx, schema1.ID)
	require.NoError(t, err)
	assert.Equal(t, schema1.ID, schema2.ID)
	assert.Equal(t, schema1.IssuerDID, schema2.IssuerDID)
	assert.Equal(t, schema1.URL, schema2.URL)
	assert.Equal(t, schema1.Type, schema2.Type)
	assert.Equal(t, schema1.Hash, schema2.Hash)
	assert.Equal(t, schema1.Attributes, schema2.Attributes)
	assert.InDelta(t, schema1.CreatedAt.UnixNano(), schema2.CreatedAt.UnixNano(), 1000)
}
