package repositories

import (
	"context"
	"testing"

	core "github.com/iden3/go-iden3-core"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
)

func TestRevoke(t *testing.T) {
	// given
	claimsRepo := NewClaims(storage.Pgx)
	idStr := "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"
	identity := &domain.Identity{
		Identifier: idStr,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	// when and then
	t.Run("should save the revocation", func(t *testing.T) {
		assert.NoError(t, claimsRepo.Revoke(context.Background(), storage.Pgx, &domain.Revocation{
			Identifier:  idStr,
			Nonce:       domain.RevNonceUint64(123),
			Version:     uint32(1),
			Status:      domain.RevPending,
			Description: "a description",
		}))
	})

	t.Run("should not save the revocation", func(t *testing.T) {
		assert.Error(t, claimsRepo.Revoke(context.Background(), storage.Pgx, &domain.Revocation{
			Identifier:  "123",
			Nonce:       domain.RevNonceUint64(123),
			Version:     uint32(1),
			Status:      domain.RevPending,
			Description: "a description",
		}))
	})
}

func TestGetByRevocationNonce(t *testing.T) {
	fixture := tests.NewFixture(storage)
	idStr := "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"
	identity := &domain.Identity{
		Identifier: idStr,
		Relay:      "relay_mock",
		Immutable:  false,
	}
	fixture.CreateIdentity(t, identity)
	idClaim, _ := uuid.NewUUID()
	fixture.CreateClaims(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      &idStr,
		Issuer:          idStr,
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: "",
		Expiration:      0,
		Version:         0,
		RevNonce:        0,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	claimsRepo := NewClaims(storage.Pgx)
	t.Run("should get revocation", func(t *testing.T) {
		did, err := core.ParseDID(idStr)
		assert.NoError(t, err)
		r, err := claimsRepo.GetByRevocationNonce(context.Background(), storage.Pgx, did, 0)
		assert.NoError(t, err)
		assert.NotNil(t, r)
	})

	t.Run("should not get revocation wrong nonce", func(t *testing.T) {
		did, err := core.ParseDID(idStr)
		assert.NoError(t, err)
		r, err := claimsRepo.GetByRevocationNonce(context.Background(), storage.Pgx, did, 1)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("should not get revocation wrong did", func(t *testing.T) {
		did, err := core.ParseDID("did:polygonid:polygon:mumbai:2qFAer2CpbpNhMCkiMCrQbUf4vXnEKPhrQmqVfnaeY")
		assert.NoError(t, err)
		r, err := claimsRepo.GetByRevocationNonce(context.Background(), storage.Pgx, did, 1)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
}
