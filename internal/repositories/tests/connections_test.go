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
