package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/jackc/pgtype"
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
		did, err := w3c.ParseDID(idStr)
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
		did, err := w3c.ParseDID(idStr)
		assert.NoError(t, err)
		r, err := claimsRepo.GetByRevocationNonce(context.Background(), storage.Pgx, did, 1)
		assert.Error(t, err)
		assert.Nil(t, r)
	})

	t.Run("should not get revocation wrong did", func(t *testing.T) {
		did, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qFAer2CpbpNhMCkiMCrQbUf4vXnEKPhrQmqVfnaeY")
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

	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qH7XAwYQzCp9VfhpNgeLtK2iCehDDrfMWUCEg5ig5")
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
	issuerDID, err := w3c.ParseDID("did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:iden3:tJUieNy7sk5PhitERHg1tgM8v1qhsDSEHVJSUF9rJ")
	require.NoError(t, err)

	vc := &verifiable.W3CCredential{
		ID: uuid.NewString(),
		CredentialSubject: map[string]any{
			"number": 1,
			"string": "some words",
		},
	}
	jsonB := &pgtype.JSONB{}
	require.NoError(t, jsonB.Set(`{"type": "BJJSignature2021", "coreClaim": "c9b2370371b7fa8b3dab2a5ba81b68382a00000000000000000000000000000002129c52957a73ea89144dc455d28e074cd7e23ae3e5bf86d4aa56d20cd60e0074da1e21d2c4d8fc28e2e3809ed51c333d68ef4dffd31508176ab84863e8fc1a0000000000000000000000000000000000000000000000000000000000000000682561f1000000006f0535010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "signature": "fb179bc43ca2c8ce4eb97549d847415bcb759f4d7c8bb3aa008700716abb2b06853349d75571fdc3018023cce9d1e6756eb102b4b44a17555d49fc8371af1300", "issuerData": {"id": "did:polygonid:polygon:mumbai:2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr", "mtp": {"siblings": [], "existence": true}, "state": {"value": "e6a67b3bcca7e424f657f41ddaae87f772f502de49d1cfe7f9abd11a4822611d", "claimsTreeRoot": "8375a237f1597b74b17f33cce0638e93a7be9175028836ae9f54f08dd2976a2f"}, "authCoreClaim": "cca3371a6cb1b715004407e325bd993c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f5287a7ac420b7c2b1b7aa28446c52df4dda6f7e4a127fbd1272d78853c4e01a3359f10f7fef6a358b83740146445dc55f143109bf1f6a090edf7d7c7b8e651c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "credentialStatus": {"id": "https://aeb5-2a0c-5a84-e10a-5200-71e6-4d79-d127-c4dd.eu.ngrok.io/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr/claims/revocation/status/0", "type": "SparseMerkleTreeProof", "revocationNonce": 0}}}`))
	c := &domain.Claim{
		ID:              uuid.New(),
		Identifier:      common.ToPointer(issuerDID.String()),
		Issuer:          issuerDID.String(),
		SchemaHash:      "ca938857241db9451ea329256b9c06e5",
		SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
		SchemaType:      "AuthBJJCredential",
		SignatureProof:  *jsonB,
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
		{
			name:     "filter.Subject should return one entry",
			filter:   ports.ClaimsFilter{Subject: userDID.String()},
			expected: 1,
		},
		{
			name:     "no mtp proof for this user",
			filter:   ports.ClaimsFilter{Subject: userDID.String(), Proofs: []verifiable.ProofType{verifiable.Iden3SparseMerkleTreeProofType}},
			expected: 0,
		},
		{
			name:     "one sig proof for this user",
			filter:   ports.ClaimsFilter{Subject: userDID.String(), Proofs: []verifiable.ProofType{verifiable.BJJSignatureProofType}},
			expected: 1,
		},
		{
			name:     "one sig proof for this user with any signature proof filter",
			filter:   ports.ClaimsFilter{Subject: userDID.String(), Proofs: []verifiable.ProofType{domain.AnyProofType}},
			expected: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims, total, err := claimsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &tc.filter)
			require.NoError(t, err)
			assert.Len(t, claims, tc.expected)
			assert.Equal(t, total, uint(len(claims)))
		})
	}
}

func TestGetAllByIssuerIDPagination(t *testing.T) {
	ctx := context.Background()
	fixture := tests.NewFixture(storage)
	issuerDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qMFKi3ou8Sd5oeHt3NquUKnPUqDMD84yvpm4pt8Hi")
	require.NoError(t, err)
	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qPnPzctLT3jEzW3aZg2yAGEeeBW6izu5znWULdNRy")
	require.NoError(t, err)

	jsonB := &pgtype.JSONB{}
	require.NoError(t, jsonB.Set(`{"type": "BJJSignature2021", "coreClaim": "c9b2370371b7fa8b3dab2a5ba81b68382a00000000000000000000000000000002129c52957a73ea89144dc455d28e074cd7e23ae3e5bf86d4aa56d20cd60e0074da1e21d2c4d8fc28e2e3809ed51c333d68ef4dffd31508176ab84863e8fc1a0000000000000000000000000000000000000000000000000000000000000000682561f1000000006f0535010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "signature": "fb179bc43ca2c8ce4eb97549d847415bcb759f4d7c8bb3aa008700716abb2b06853349d75571fdc3018023cce9d1e6756eb102b4b44a17555d49fc8371af1300", "issuerData": {"id": "did:polygonid:polygon:mumbai:2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr", "mtp": {"siblings": [], "existence": true}, "state": {"value": "e6a67b3bcca7e424f657f41ddaae87f772f502de49d1cfe7f9abd11a4822611d", "claimsTreeRoot": "8375a237f1597b74b17f33cce0638e93a7be9175028836ae9f54f08dd2976a2f"}, "authCoreClaim": "cca3371a6cb1b715004407e325bd993c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000f5287a7ac420b7c2b1b7aa28446c52df4dda6f7e4a127fbd1272d78853c4e01a3359f10f7fef6a358b83740146445dc55f143109bf1f6a090edf7d7c7b8e651c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", "credentialStatus": {"id": "https://aeb5-2a0c-5a84-e10a-5200-71e6-4d79-d127-c4dd.eu.ngrok.io/v1/did%3Apolygonid%3Apolygon%3Amumbai%3A2qL68in3FNbimFK6gka8hPZz475z31nqPJdqBeTsQr/claims/revocation/status/0", "type": "SparseMerkleTreeProof", "revocationNonce": 0}}}`))

	createdAt := time.Now().Add(-24 * time.Hour)
	for i := 0; i < 100; i++ {
		c := &domain.Claim{
			ID:              uuid.New(),
			Identifier:      common.ToPointer(issuerDID.String()),
			Issuer:          issuerDID.String(),
			SchemaHash:      "ca938857241db9451ea329256b9c06e5",
			SchemaURL:       "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld",
			SchemaType:      "AuthBJJCredential",
			SignatureProof:  *jsonB,
			OtherIdentifier: userDID.String(),
			HIndex:          fmt.Sprintf("%d", rand.Int()),
			CreatedAt:       createdAt,
		}
		require.NoError(t, c.Data.Set(&verifiable.W3CCredential{
			ID: uuid.NewString(),
			CredentialSubject: map[string]any{
				"number": 1,
				"string": "some words",
			},
		}))

		_ = fixture.CreateClaim(t, c)
		createdAt = createdAt.Add(time.Second)
	}

	claimsRepo := repositories.NewClaims()

	type expected struct {
		total     uint
		resultLen uint
	}

	type testConfig struct {
		name     string
		filter   ports.ClaimsFilter
		expected expected
	}
	for _, tc := range []testConfig{
		{
			name: "If page is nil, return all",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 100,
				Page:       nil,
			},
			expected: expected{
				total:     100,
				resultLen: 100,
			},
		},
		{
			name: "Return first page of 100",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 100,
				Page:       common.ToPointer(uint(1)),
			},
			expected: expected{
				total:     100,
				resultLen: 100,
			},
		},
		{
			name: "Return first page of 25",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 25,
				Page:       common.ToPointer(uint(1)),
			},
			expected: expected{
				total:     100,
				resultLen: 25,
			},
		},
		{
			name: "Return first page of 25",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 25,
				Page:       common.ToPointer(uint(1)),
			},
			expected: expected{
				total:     100,
				resultLen: 25,
			},
		},
		{
			name: "Return 4 page of 33",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 33,
				Page:       common.ToPointer(uint(4)),
			},
			expected: expected{
				total:     100,
				resultLen: 1,
			},
		},
		{
			name: "Return 100 page of 1",
			filter: ports.ClaimsFilter{
				Subject:    userDID.String(),
				MaxResults: 1,
				Page:       common.ToPointer(uint(100)),
			},
			expected: expected{
				total:     100,
				resultLen: 1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims, total, err := claimsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &tc.filter)
			require.NoError(t, err)
			assert.Len(t, claims, int(tc.expected.resultLen))
			assert.Equal(t, total, tc.expected.total)

			// Let's check ids, etc...
			all := tc.filter
			all.Page = nil
			allClaims, total, err := claimsRepo.GetAllByIssuerID(ctx, storage.Pgx, *issuerDID, &all)
			require.NoError(t, err)

			var from uint = 0
			to := total
			if tc.filter.Page != nil {
				from = (*tc.filter.Page - 1) * tc.filter.MaxResults
				to = from + tc.filter.MaxResults
				if to >= total {
					to = total - 1
				}
			}
			for i := from; i < to; i++ {
				assert.Equal(t, allClaims[i].ID, claims[i].ID, "iteration: %d", i)
			}
		})
	}
}

func TestGetClaimsIssuedForUserID(t *testing.T) {
	ctx := context.Background()
	fixture := tests.NewFixture(storage)
	didStr := "did:polygonid:polygon:mumbai:2qKLWeRi6Tk23SmFpRKHvKFf2MmrocJYxwAD1MwhYw"
	schemaStore := repositories.NewSchema(*storage)
	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	require.NoError(t, err)
	linkStore := repositories.NewLink(*storage)

	schemaID := insertSchemaForLink(ctx, didStr, schemaStore, t)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	tomorrow := time.Now().Add(24 * time.Hour)
	nextWeek := time.Now().Add(7 * 24 * time.Hour)

	link := domain.NewLink(*did, common.ToPointer[int](10), &tomorrow, schemaID, &nextWeek, true, false, domain.CredentialSubject{})
	link.MaxIssuance = common.ToPointer(100)

	linkID, err := linkStore.Save(ctx, storage.Pgx, link)
	require.NoError(t, err)
	assert.NotNil(t, linkID)

	idClaim, _ := uuid.NewUUID()
	HIndex := uuid.New().String()

	userDID, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qP8KN3KRwBi37jB2ENXrWxhTo3pefaU5u5BFPbjYo")
	require.NoError(t, err)

	userDIDWithCeroClaims, err := w3c.ParseDID("did:polygonid:polygon:mumbai:2qHLU5GYftBHunAEh5PrBifeJiEVujh9Ybzukh7Nhy")
	require.NoError(t, err)

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
		userDID  w3c.DID
		expected int
	}

	claimsRepo := repositories.NewClaims()

	for _, tc := range []testConfig{
		{
			name:     "should return 1",
			userDID:  *userDID,
			expected: 1,
		},
		{
			name:     "should return 0",
			userDID:  *userDIDWithCeroClaims,
			expected: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := claimsRepo.GetClaimsIssuedForUser(ctx, storage.Pgx, *did, tc.userDID, link.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, len(claims))
		})
	}
}
