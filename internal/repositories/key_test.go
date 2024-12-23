package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

func TestKey_Save(t *testing.T) {
	keyRepository := NewKey(*storage)

	ctx := context.Background()
	did := randomDID(t)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", did.String(), "BJJ")
	assert.NoError(t, err)

	t.Run("should save a new key", func(t *testing.T) {
		key := domain.Key{
			ID:        uuid.New(),
			IssuerDID: domain.KeyCoreDID(did),
			PublicKey: "publicKey",
			Name:      "name",
		}
		id, err := keyRepository.Save(context.Background(), storage.Pgx, &key)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
		assert.Equal(t, key.ID, id)
	})

	t.Run("should get an error", func(t *testing.T) {
		key := domain.Key{
			ID:        uuid.New(),
			IssuerDID: domain.KeyCoreDID(did),
			PublicKey: "publicKey",
			Name:      "name_1",
		}

		id, err := keyRepository.Save(context.Background(), storage.Pgx, &key)
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, id)
		assert.Equal(t, key.ID, id)

		key2 := domain.Key{
			ID:        uuid.New(),
			IssuerDID: domain.KeyCoreDID(did),
			PublicKey: "publicKey2",
			Name:      "name_1",
		}

		id, err = keyRepository.Save(context.Background(), storage.Pgx, &key2)
		require.Error(t, err)
		require.Equal(t, uuid.Nil, id)
	})
}

func TestKey_GetByPublicKey(t *testing.T) {
	keyRepository := NewKey(*storage)
	ctx := context.Background()
	did := randomDID(t)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", did.String(), "BJJ")
	assert.NoError(t, err)

	key := domain.Key{
		ID:        uuid.New(),
		IssuerDID: domain.KeyCoreDID(did),
		PublicKey: "publicKey",
		Name:      "name" + uuid.New().String(),
	}
	id, err := keyRepository.Save(context.Background(), storage.Pgx, &key)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
	assert.Equal(t, key.ID, id)

	t.Run("should get the key by public key", func(t *testing.T) {
		keyFromDatabase, err := keyRepository.GetByPublicKey(ctx, did, key.PublicKey)
		require.NoError(t, err)
		assert.Equal(t, key.ID, keyFromDatabase.ID)
		assert.Equal(t, key.IssuerDID, keyFromDatabase.IssuerDID)
		assert.Equal(t, key.PublicKey, keyFromDatabase.PublicKey)
		assert.Equal(t, key.Name, keyFromDatabase.Name)
	})

	t.Run("should get an error - ErrKeyNotFound", func(t *testing.T) {
		keyFromDatabase, err := keyRepository.GetByPublicKey(ctx, did, "wrong public key")
		assert.Error(t, err)
		assert.Nil(t, keyFromDatabase)
		assert.True(t, errors.Is(err, ErrKeyNotFound))
	})
}
