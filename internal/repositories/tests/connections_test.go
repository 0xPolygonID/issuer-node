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
	issuerDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qCp9Tx4x5hzchym1dZXtBpwRQsH7HXe7GcbvskoRn")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm")
	require.NoError(t, err)

	conn := &domain.Connection{
		ID:        uuid.New(),
		UserDID:   *userDID,
		IssuerDID: *issuerDID,
	}

	conn2 := &domain.Connection{
		ID:        uuid.New(),
		UserDID:   *userDID,
		IssuerDID: *issuerDID,
	}
	t.Run("should save or update the connection", func(t *testing.T) {
		connID, err := connectionsRepo.Save(context.Background(), storage.Pgx, conn)
		assert.NoError(t, err)
		assert.Equal(t, conn.ID.String(), connID.String())
		connID2, err := connectionsRepo.Save(context.Background(), storage.Pgx, conn2) // updating connection
		assert.NoError(t, err)
		assert.NotEqual(t, conn2.ID, connID2) // checking that the connections is being updated and no ID is modified on conflict
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
		assert.Error(t, connectionsRepo.Delete(context.Background(), storage.Pgx, uuid.New(), *issuerDID))
	})

	t.Run("should get an error, deleting non existing issuer connection", func(t *testing.T) {
		assert.Error(t, connectionsRepo.Delete(context.Background(), storage.Pgx, uuid.New(), *userDID))
	})

	t.Run("should delete an existing connection", func(t *testing.T) {
		assert.NoError(t, connectionsRepo.Delete(context.Background(), storage.Pgx, conn, *issuerDID))
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
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *userDID, "")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 0)
	})

	t.Run("should get 1 connection for a the given issuerDID and no query", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, "")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, just beginning", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, "did:")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, full did", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, "did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, part of did", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, "did:polygonid:polygon:mumbai:2qH7XAw")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 1)
	})

	t.Run("should get 0 connections for a the given issuerDID and non existing userDID", func(t *testing.T) {
		conns, err := connectionsRepo.GetAllByIssuerID(context.Background(), storage.Pgx, *issuerDID, "did:polygonid:polygon:mumbai:2qH7XAwnonexisting")
		require.NoError(t, err)
		assert.Equal(t, len(conns), 0)
	})
}

func TestDeleteConnectionCredentials(t *testing.T) {
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

	fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246564",
	})

	fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246563",
	})

	t.Run("should delete all the credentials", func(t *testing.T) {
		assert.NoError(t, connectionsRepo.DeleteCredentials(context.Background(), storage.Pgx, conn, *issuerDID))
	})

	t.Run("should return no error for non existing connection", func(t *testing.T) {
		assert.NoError(t, connectionsRepo.DeleteCredentials(context.Background(), storage.Pgx, uuid.New(), *issuerDID))
	})
}

func TestGetByUserID(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)
	userDID2, err := core.ParseDID("did:polygonid:polygon:mumbai:2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr")
	require.NoError(t, err)

	_ = fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246564",
	})

	fixture.CreateClaim(t, &domain.Claim{
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		OtherIdentifier: userDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246563",
	})

	t.Run("should get an error, no connection for the given userID", func(t *testing.T) {
		_, err := connectionsRepo.GetByUserID(context.Background(), storage.Pgx, *issuerDID, *userDID2)
		assert.Error(t, err)
	})
	t.Run("should get a connection for the given userID", func(t *testing.T) {
		conn, err := connectionsRepo.GetByUserID(context.Background(), storage.Pgx, *issuerDID, *userDID)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	})
}
