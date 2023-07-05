package tests

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestGetSchema(t *testing.T) {
	rand.NewSource(time.Now().Unix())
	ctx := context.Background()
	store := repositories.NewSchema(*storage)
	did := core.DID{}
	// Create a schemaHash
	i := &big.Int{}
	i.SetInt64(rand.Int63())

	require.NoError(t, did.SetString("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"))
	schema1 := &domain.Schema{
		ID:          uuid.New(),
		IssuerDID:   did,
		URL:         "https://an.url.org/index.html",
		Type:        "schemaType",
		Hash:        core.NewSchemaHashFromInt(i),
		Words:       domain.SchemaWords{"field1", "field2", "fieldn"},
		CreatedAt:   time.Now(),
		Title:       "some title",
		Description: "some description",
		Version:     "1.0.0",
	}
	require.NoError(t, store.Save(ctx, schema1))

	schema2, err := store.GetByID(ctx, did, schema1.ID)
	require.NoError(t, err)
	assert.Equal(t, schema1.ID, schema2.ID)
	assert.Equal(t, schema1.IssuerDID, schema2.IssuerDID)
	assert.Equal(t, schema1.URL, schema2.URL)
	assert.Equal(t, schema1.Type, schema2.Type)
	assert.Equal(t, schema1.Hash, schema2.Hash)
	assert.Equal(t, domain.SchemaWords{"schemaType", "field1", "field2", "fieldn"}, schema2.Words)
	assert.InDelta(t, schema1.CreatedAt.UnixMilli(), schema2.CreatedAt.UnixMilli(), 10)
	assert.Equal(t, schema1.Title, schema2.Title)
	assert.Equal(t, schema1.Description, schema2.Description)
	assert.Equal(t, schema1.Version, schema2.Version)
}

func TestCreateSchema(t *testing.T) {
	rand.NewSource(time.Now().Unix())
	ctx := context.Background()
	store := repositories.NewSchema(*storage)
	did := core.DID{}
	// Create a schemaHash
	i := &big.Int{}
	i.SetInt64(rand.Int63())

	require.NoError(t, did.SetString("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"))
	schema1 := &domain.Schema{
		ID:          uuid.New(),
		IssuerDID:   did,
		URL:         "https://an.url.org/index.html",
		Type:        "schemaType",
		Hash:        core.NewSchemaHashFromInt(i),
		Words:       domain.SchemaWords{"field1", "field2", "fieldn"},
		CreatedAt:   time.Now(),
		Title:       "some title",
		Description: "some description",
		Version:     uuid.NewString(),
	}

	require.NoError(t, store.Save(ctx, schema1))
	assert.Error(t, store.Save(ctx, schema1), "cannot have duplicated schemas with the same version for the same issuer and type")

	schema2 := schema1
	schema2.Version = uuid.NewString()
	schema2.ID = uuid.New()
	assert.NoError(t, store.Save(ctx, schema2))
}

func TestGetAllFullTextSearch(t *testing.T) {
	rand.NewSource(time.Now().Unix())
	ctx := context.Background()
	// Need an isolated DB here to avoid other tests side effects
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}
	storage, teardown, err := tests.NewTestStorage(&config.Configuration{Database: config.Database{URL: conn}})
	require.NoError(t, err)
	defer teardown()

	store := repositories.NewSchema(*storage)
	did := core.DID{}
	require.NoError(t, did.SetString("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"))
	insertSchemaGetAllData(t, ctx, did, store)

	type expected struct {
		collection []domain.Schema
	}
	type testConfig struct {
		name     string
		query    *string
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name:  "Nil query. Expect all entries",
			query: nil,
			expected: expected{
				collection: []domain.Schema{{
					Type:  "nicePeopleAtWork",
					Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
				}, {
					Type:  "age",
					Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
				}},
			},
		},
		{
			name:  "Empty query. Expect all entries",
			query: common.ToPointer(""),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "nicePeopleAtWork",
					Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
				}, {
					Type:  "age",
					Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
				}},
			},
		},
		{
			name:  "Exact math for schema type. Expect one",
			query: common.ToPointer("nicePeopleAtWork"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "nicePeopleAtWork",
					Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
				}},
			},
		},
		{
			name:  "Exact math for schema type in lower case . Expect one",
			query: common.ToPointer("nicepeopleatwork"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "nicePeopleAtWork",
					Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
				}},
			},
		},
		{
			name:  "partial match for schema type beginning. Expect one",
			query: common.ToPointer("nicepeoplea"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "nicePeopleAtWork",
					Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
				}},
			},
		},
		{
			name:  "Exact match attributes",
			query: common.ToPointer("younger than eighteen"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "age",
					Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
				}},
			},
		},
		{
			name:  "partial match attributes",
			query: common.ToPointer("eighteen"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "age",
					Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
				}},
			},
		},
		{
			name:  "partial match attributes, middle of the word",
			query: common.ToPointer("eighte"),
			expected: expected{
				collection: []domain.Schema{{
					Type:  "age",
					Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
				}},
			},
		},
		{
			name:  "2 attributes from different records",
			query: common.ToPointer("younger smart"),
			expected: expected{
				collection: []domain.Schema{
					{
						Type:  "nicePeopleAtWork",
						Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
					},
					{
						Type:  "age",
						Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
					},
				},
			},
		},
		{
			name:  "2 attributes from different records, partial match",
			query: common.ToPointer("young people"),
			expected: expected{
				collection: []domain.Schema{
					{
						Type:  "nicePeopleAtWork",
						Words: domain.SchemaWords{"nicePeopleAtWork", "friendly", "helper", "empathic", "smart"},
					},
					{
						Type:  "age",
						Words: domain.SchemaWords{"age", "younger than eighteen", "older than eighteen"},
					},
				},
			},
		},
		{
			name:  "Something including did:******",
			query: common.ToPointer("some words and did:polygonid:polygon:mumbai:2qE1BZ"),
			expected: expected{
				collection: []domain.Schema{},
			},
		},

		// TODO: Add partial like tests
	} {
		t.Run(tc.name, func(t *testing.T) {
			collection, err := store.GetAll(ctx, did, tc.query)
			require.NoError(t, err)
			require.Len(t, collection, len(tc.expected.collection))
			for i := range collection {
				assert.Equal(t, tc.expected.collection[i].Words, collection[i].Words)
			}
		})
	}
}

func insertSchemaGetAllData(t *testing.T, ctx context.Context, did core.DID, store ports.SchemaRepository) {
	t.Helper()
	data := []struct {
		typ        string
		attributes domain.SchemaWords
	}{
		{typ: "age", attributes: domain.SchemaWords{"younger than eighteen", "older than eighteen"}},
		{typ: "nicePeopleAtWork", attributes: domain.SchemaWords{"friendly", "helper", "empathic", "smart"}},
	}

	for i, d := range data {
		s := &domain.Schema{
			ID:        uuid.New(),
			IssuerDID: did,
			URL:       fmt.Sprintf("url is not important in this test but need to be unique %d", i),
			Type:      d.typ,
			Words:     d.attributes,
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.Save(ctx, s))
		time.Sleep(2 * time.Millisecond)
	}
}
