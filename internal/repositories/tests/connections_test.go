package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestSave(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qCp9Tx4x5hzchym1dZXtBpwRQsH7HXe7GcbvskoRn")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm")
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
		ctx := context.Background()
		connID, err := connectionsRepo.Save(ctx, storage.Pgx, conn)
		assert.NoError(t, err)
		assert.Equal(t, conn.ID.String(), connID.String())
		connID2, err := connectionsRepo.Save(ctx, storage.Pgx, conn2) // updating connection
		assert.NoError(t, err)
		assert.NotEqual(t, conn2.ID, connID2) // checking that the connections is being updated and no ID is modified on conflict
	})
}

func TestUpdatePushToken(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qDLHs1n3c9oHxEPkgCMGfDjY4V37Xv8KztkZcpG1i")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm")
	require.NoError(t, err)

	conn := &domain.Connection{
		ID:        uuid.New(),
		UserDID:   *userDID,
		IssuerDID: *issuerDID,
		UserDoc:   json.RawMessage(`{"id": "did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm", "service": [{"id": "did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm#push", "type": "push-notification", "metadata": {"devices": [{"alg": "RSA-OAEP-512", "ciphertext": "someToken"}]}, "serviceEndpoint": "https://someURL.com/api/v1"}], "@context": ["https://www.w3.org/ns/did/v1"]}`),
	}

	conn2 := &domain.Connection{
		ID:        conn.ID,
		UserDID:   *userDID,
		IssuerDID: *issuerDID,
		UserDoc:   json.RawMessage(`{"id": "did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm", "service": [{"id": "did:polygonid:polygon:mumbai:2qHgCmGW1wDH5ShTH94SssR4eN8XW4xyHLfop2Qoqm#push", "type": "push-notification", "metadata": {"devices": [{"alg": "RSA-OAEP-512", "ciphertext": "someToken2"}]}, "serviceEndpoint": "https://someURL.com/api/v1"}], "@context": ["https://www.w3.org/ns/did/v1"]}`),
	}
	t.Run("should save or update the connection", func(t *testing.T) {
		ctx := context.Background()
		connID, err := connectionsRepo.Save(ctx, storage.Pgx, conn)
		require.NoError(t, err)
		assert.Equal(t, conn.ID.String(), connID.String())
		connID2, err := connectionsRepo.Save(ctx, storage.Pgx, conn2) // updating connection
		require.NoError(t, err)
		assert.Equal(t, conn2.ID, connID2)
		connDB, err := connectionsRepo.GetByUserID(ctx, storage.Pgx, *issuerDID, *userDID)
		require.NoError(t, err)
		assert.Equal(t, conn2.ID, connDB.ID)
		assert.Equal(t, conn2.UserDoc, connDB.UserDoc)
	})
}

func TestSaveUserAuthentication(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:polygonid:ethereum:main:2qKDJmySKNi4GD4vYdqfLb37MSTSijg77NoRZaKfDX")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:ethereum:main:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	connID := fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})
	sessionID := uuid.New()

	require.NoError(t, connectionsRepo.SaveUserAuthentication(context.Background(), storage.Pgx, connID, sessionID, time.Now()))

	connDB, err := connectionsRepo.GetByUserSessionID(context.Background(), storage.Pgx, sessionID)
	require.NoError(t, err)

	assert.Equal(t, connDB.ID, connID)
	assert.Equal(t, connDB.IssuerDID.String(), issuerDID.String())
	assert.Equal(t, connDB.UserDID.String(), userDID.String())
}

func TestDelete(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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

func TestConnectionsGetAllByIssuerID(t *testing.T) {
	ctx := context.Background()
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *userDID, &ports.NewGetAllConnectionsRequest{Query: ""})
		require.NoError(t, err)
		assert.Equal(t, 0, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and no query", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: ""})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, just beginning", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "did:"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, full did", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and valid query, part of did", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "did:polygonid:polygon:mumbai:2qH7XAw"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and a query with some chars in the middle of a string", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "H7XAw"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 1 connection for a the given issuerDID and a query with some chars in the middle of a string and other words", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "H7XAw other words"})
		require.NoError(t, err)
		assert.Equal(t, 1, len(conns))
	})

	t.Run("should get 0 connections for a the given issuerDID and non existing userDID", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: "did:polygonid:polygon:mumbai:2qH7XAwnonexisting"})
		require.NoError(t, err)
		assert.Equal(t, 0, len(conns))
	})
}

func TestGetAllWithCredentialsByIssuerID(t *testing.T) {
	ctx := context.Background()
	connectionsRepo := repositories.NewConnections()

	fixture := tests.NewFixture(storage)
	idStr := "did:polygonid:polygon:mumbai:2qEinAT1jt9vfDfEwdjdD4B3vGJxMAVjgK2yvvKij4"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture.CreateIdentity(t, identity)
	issuerDID, err := w3c.ParseDID(idStr)
	require.NoError(t, err)

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qNtJm8v3c8b7XjQtAtSvAbudnUAfzsjHFqRnyYDq7")
	require.NoError(t, err)
	userDID2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFjTM4kX3J6AYzHBY1Q3ztnxv1UfNaaNUGw8TKo4N")
	require.NoError(t, err)
	userDID3, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qMTdi9CkqE8ihMn7qtp61QCdiKvfo2Ttx9a5TMDSt")
	require.NoError(t, err)

	connNoCredentials := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	connWithCredentials := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID2,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	connLatestIssuedCredentials := fixture.CreateConnection(t, &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID3,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})
	orderIDs := []uuid.UUID{connLatestIssuedCredentials, connWithCredentials, connNoCredentials}

	fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246562",
		OtherIdentifier: userDID2.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        1234,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		CreatedAt:       time.Now(),
	})

	fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246563",
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID3.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        1234,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		CreatedAt:       time.Now().Add(time.Hour),
	})

	fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		HIndex:          "20060639968773997271173557722944342103398298534714534718204282267207714246566",
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID3.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        1234,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		CreatedAt:       time.Now().Add(-24 * time.Hour),
	})

	t.Run("should get 3 connections for the given issuerDID", func(t *testing.T) {
		conns, _, err := connectionsRepo.GetAllWithCredentialsByIssuerID(ctx, storage.Pgx, *issuerDID, &ports.NewGetAllConnectionsRequest{Query: ""})
		require.NoError(t, err)
		require.Equal(t, len(orderIDs), len(conns))

		for i := range conns {
			assert.Equal(t, orderIDs[i].String(), conns[i].ID.String())
		}
	})
}

func TestDeleteConnectionCredentials(t *testing.T) {
	connectionsRepo := repositories.NewConnections()
	fixture := tests.NewFixture(storage)

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)
	userDID2, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr")
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
