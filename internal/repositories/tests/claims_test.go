package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

func TestRevoke(t *testing.T) {
	// given
	claimsRepo := repositories.NewClaims()
	idStr := "did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ"
	identity := &domain.Identity{
		Identifier: idStr,
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
	idStr := "did:polygonid:polygon:mumbai:2qHtzzxS7uazdumnyZEdf74CNo3MptdW6ytxxwbPMW"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture.CreateIdentity(t, identity)
	idClaim, _ := uuid.NewUUID()
	fixture.CreateClaim(t, &domain.Claim{
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

	claimsRepo := repositories.NewClaims()
	t.Run("should get revocation", func(t *testing.T) {
		did, err := core.ParseDID(idStr)
		assert.NoError(t, err)
		c, err := claimsRepo.GetByRevocationNonce(context.Background(), storage.Pgx, did, 0)
		assert.NoError(t, err)
		assert.NotNil(t, c)
		coreClaimValue, err := c.CoreClaim.Value()
		assert.NoError(t, err)
		assert.Equal(t, idClaim, c.ID)
		assert.Equal(t, &idStr, c.Identifier)
		assert.Equal(t, "ca938857241db9451ea329256b9c06e5", c.SchemaHash)
		assert.Equal(t, "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld", c.SchemaURL)
		assert.Equal(t, "AuthBJJCredential", c.SchemaType)
		assert.Equal(t, "", c.OtherIdentifier)
		assert.Equal(t, int64(0), c.Expiration)
		assert.Equal(t, uint32(0), c.Version)
		assert.Equal(t, domain.RevNonceUint64(0), c.RevNonce)
		assert.Equal(t, `["0","0","0","0","0","0","0","0"]`, coreClaimValue)

		assert.Nil(t, c.Status)
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

func TestRevokeNonce(t *testing.T) {
	// given
	claimsRepo := repositories.NewClaims()
	idStr := "did:polygonid:polygon:mumbai:2qNWrZ4Z7iZPvDusp32sWXGMHvAL9RoTqgPEEXvS9q"
	identity := &domain.Identity{
		Identifier: idStr,
	}
	fixture := tests.NewFixture(storage)
	fixture.CreateIdentity(t, identity)

	// when and then
	t.Run("should save the revocation nonce", func(t *testing.T) {
		assert.NoError(t, claimsRepo.RevokeNonce(context.Background(), storage.Pgx, &domain.Revocation{
			Identifier:  idStr,
			Nonce:       domain.RevNonceUint64(123),
			Version:     uint32(1),
			Status:      domain.RevPending,
			Description: "a description",
		}))
	})

	t.Run("should not save the revocation", func(t *testing.T) {
		assert.Error(t, claimsRepo.RevokeNonce(context.Background(), storage.Pgx, &domain.Revocation{
			Identifier:  "123",
			Nonce:       domain.RevNonceUint64(123),
			Version:     uint32(1),
			Status:      domain.RevPending,
			Description: "a description",
		}))
	})
}

func TestGetAllByConnectionAndIssuerID(t *testing.T) {
	fixture := tests.NewFixture(storage)

	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
	require.NoError(t, err)

	_ = fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        0,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
	})

	_ = fixture.CreateClaim(t, &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		HIndex:          uuid.NewString(),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e2",
		SchemaURL:       "https://raw.githubusercontent.com/iden2/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential2",
		OtherIdentifier: userDID.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        0,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		Revoked:         true,
	})

	conn := fixture.CreateConnection(t, &domain.Connection{
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  nil,
		UserDoc:    nil,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	})

	claimsRepo := repositories.NewClaims()
	t.Run("should get one claim", func(t *testing.T) {
		r, err := claimsRepo.GetNonRevokedByConnectionAndIssuerID(context.Background(), storage.Pgx, conn, *issuerDID)
		assert.NoError(t, err)
		assert.Equal(t, len(r), 1)
	})

	t.Run("should get no claims, issuerDID not found", func(t *testing.T) {
		r, err := claimsRepo.GetNonRevokedByConnectionAndIssuerID(context.Background(), storage.Pgx, conn, *userDID)
		assert.NoError(t, err)
		assert.Equal(t, len(r), 0)
	})

	t.Run("should get no claims, connID not found", func(t *testing.T) {
		r, err := claimsRepo.GetNonRevokedByConnectionAndIssuerID(context.Background(), storage.Pgx, uuid.New(), *issuerDID)
		assert.NoError(t, err)
		assert.Equal(t, len(r), 0)
	})
}

func TestGetAllByIssuerID(t *testing.T) {
	ctx := context.Background()

	fixture := tests.NewFixture(storage)
	issuerDID, err := core.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := core.ParseDID("did:iden3:tJUieNy7sk5PhitERHg1tgM8v1qhsDSEHVJSUF9rJ")
	require.NoError(t, err)

	vc := &verifiable.W3CCredential{
		ID: uuid.NewString(),
		CredentialSubject: map[string]any{
			"number": 1,
			"string": "some words",
		},
	}
	c := &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID.String(),
		HIndex:          fmt.Sprintf("%d", rand.Int()),
	}
	require.NoError(t, c.Data.Set(vc))

	_ = fixture.CreateClaim(t, c)

	claimsRepo := repositories.NewClaims()

	type testConfig struct {
		name     string
		filter   ports.ClaimsFilter
		expected int
	}
	for _, tc := range []testConfig{
		{
			name:     "filter.QueryField not found",
			filter:   ports.ClaimsFilter{QueryField: "unknown key", QueryFieldValue: "value"},
			expected: 0,
		},
		{
			name:     "filter.QueryField exists, value does not exists ",
			filter:   ports.ClaimsFilter{QueryField: "number", QueryFieldValue: "1"},
			expected: 1,
		},
		{
			name:     "filter.QueryField exists, value exists",
			filter:   ports.ClaimsFilter{QueryField: "number", QueryFieldValue: "1"},
			expected: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := claimsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &tc.filter)
			require.NoError(t, err)
			assert.Len(t, claims, tc.expected)
		})
	}
}

func TestGetClaimsIssuedByUserID(t *testing.T) {
	ctx := context.Background()
	fixture := tests.NewFixture(storage)
	didStr := "did:polygonid:polygon:mumbai:2qKLWeRi6Tk23SmFpRKHvKFf2MmrocJYxwAD1MwhYw"
	schemaStore := repositories.NewSchema(*storage)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier) VALUES ($1)", didStr)
	require.NoError(t, err)
	linkStore := repositories.NewLink(*storage)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)
	did := core.DID{}
	require.NoError(t, did.SetString(didStr))

	tomorrow := time.Now().Add(24 * time.Hour)
	nextWeek := time.Now().Add(7 * 24 * time.Hour)
	//past := time.Now().Add(-100 * 24 * time.Hour)

	link := domain.NewLink(did, common.ToPointer[int](10), &tomorrow, schemaID, &nextWeek, true, false)
	link.MaxIssuance = common.ToPointer(100)

	linkID, err := linkStore.Save(ctx, storage.Pgx, link)
	require.NoError(t, err)
	assert.NotNil(t, linkID)

	idClaim, _ := uuid.NewUUID()
	HIndex := uuid.New().String()

	userDID := core.DID{}
	require.NoError(t, userDID.SetString("did:polygonid:polygon:mumbai:2qP8KN3KRwBi37jB2ENXrWxhTo3pefaU5u5BFPbjYo"))

	userDIDWithCeroClaims := core.DID{}
	require.NoError(t, userDID.SetString("did:polygonid:polygon:mumbai:2qHLU5GYftBHunAEh5PrBifeJiEVujh9Ybzukh7Nhy"))

	idClaimInserted := fixture.CreateClaim(t, &domain.Claim{
		ID:              idClaim,
		Identifier:      common.ToPointer(did.String()),
		Issuer:          did.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		OtherIdentifier: userDID.String(),
		Expiration:      0,
		Version:         0,
		RevNonce:        0,
		CoreClaim:       domain.CoreClaim{},
		Status:          nil,
		HIndex:          HIndex,
		LinkID:          linkID,
	})

	assert.Equal(t, idClaim, idClaimInserted)

	type testConfig struct {
		name     string
		userDID  core.DID
		expected int
	}

	claimsRepo := repositories.NewClaims()

	for _, tc := range []testConfig{
		{
			name:     "should return 1",
			userDID:  userDID,
			expected: 1,
		},
		{
			name:     "should return 0",
			userDID:  userDIDWithCeroClaims,
			expected: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := claimsRepo.GetClaimsIssuedForUser(ctx, storage.Pgx, &did, &tc.userDID, link.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, len(claims))
		})
	}

}
