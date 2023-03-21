package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestSave(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	// when and then
	t.Run("should save the connection", func(t *testing.T) {
		_, err := connectionsRepo.Save(context.Background(), storage.Pgx, &domain.Connection{
			UserDID:   *userDID,
			IssuerDID: *issuerDID,
		})
		assert.NoError(t, err)
	})
}

func TestDelete(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	conn := fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	t.Run("should get an error, deleting non existing connection", func(t *testing.T) {
		assert.Error(t, connectionsRepo.Delete(context.Background(), storage.Pgx, uuid.New()))
	})

	t.Run("should delete an existing connection", func(t *testing.T) {
		assert.NoError(t, connectionsRepo.Delete(context.Background(), storage.Pgx, conn))
	})
}

func TestGetConnections(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	_ = fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	t.Run("should get 0 connections for a non existing issuerDID", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *userDID, nil)
		require.NoError(t, err)
		assert.Equal(t, len(conns), 0)
	})

	t.Run("should get 1 connection for a the given issuerDID and no query", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, nil)
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, just beginning", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, common.ToPointer("did:"))
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, full did", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, common.ToPointer("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5"))
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, part of did", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, common.ToPointer("did:polygonid:polygon:mumbai:2qH7XAw"))
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 0 connections for a the given issuerDID and non existing userDID", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, common.ToPointer("did:polygonid:polygon:mumbai:2qH7XAwnonexisting"))
		require.NoError(t, err)
		assert.Equal(t, len(conns), 0)
	})
}
