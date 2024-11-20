package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

func TestSaveVerificationQuery(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:x9b7eWa8k5rTuDBiWPHou4AvCAd1XTRfxx2uQCni8"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should save the verification", func(t *testing.T) {
		credentialSubject := pgtype.JSONB{}
		err = credentialSubject.Set(`{
		"birthday": {
			"$eq": 19791109
		}
		}`)

		credentialSubject2 := pgtype.JSONB{}
		err = credentialSubject2.Set(` {"position": {"$eq": 1}}`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           1,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
					AllowedIssuers:    []string{"issuer1", "issuer2"},
					CredentialType:    "KYCAgeCredential",
					CredentialSubject: credentialSubject,
				},
				{
					ID:                uuid.New(),
					ScopeID:           2,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "ipfs://QmaBJzpoYT2CViDx5ShJiuYLKXizrPEfXo8JqzrXCvG6oc",
					AllowedIssuers:    []string{"*"},
					CredentialType:    "TestInteger01",
					CredentialSubject: credentialSubject2,
				},
			},
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)
	})
}

func TestGetVerification(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xBdqiqz3yVT79NEAuNaqKSDZ6a5V6q8Ph66i5d2tT"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should get the verification", func(t *testing.T) {
		credentialSubject := pgtype.JSONB{}
		err = credentialSubject.Set(`{"birthday": {"$eq": 19791109}}`)
		credentialSubject2 := pgtype.JSONB{}
		err = credentialSubject2.Set(`{"position": {"$eq": 1}}`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           1,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
					AllowedIssuers:    []string{"issuer1", "issuer2"},
					CredentialType:    "KYCAgeCredential",
					CredentialSubject: credentialSubject,
				},
				{
					ID:                uuid.New(),
					ScopeID:           2,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "ipfs://QmaBJzpoYT2CViDx5ShJiuYLKXizrPEfXo8JqzrXCvG6oc",
					AllowedIssuers:    []string{"*"},
					CredentialType:    "TestInteger01",
					CredentialSubject: credentialSubject2,
				},
			},
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)

		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, verificationQueryID)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryFromDB.ID)
		assert.Equal(t, verificationQuery.ChainID, verificationQueryFromDB.ChainID)
		assert.Equal(t, verificationQuery.SkipCheckRevocation, verificationQueryFromDB.SkipCheckRevocation)
		assert.Equal(t, verificationQuery.Scopes[0].ID, verificationQueryFromDB.Scopes[0].ID)
		assert.Equal(t, verificationQuery.Scopes[0].ScopeID, verificationQueryFromDB.Scopes[0].ScopeID)
		assert.Equal(t, verificationQuery.Scopes[0].CircuitID, verificationQueryFromDB.Scopes[0].CircuitID)
		assert.Equal(t, verificationQuery.Scopes[0].Context, verificationQueryFromDB.Scopes[0].Context)
		assert.Equal(t, verificationQuery.Scopes[0].AllowedIssuers, verificationQueryFromDB.Scopes[0].AllowedIssuers)
		assert.Equal(t, verificationQuery.Scopes[0].CredentialType, verificationQueryFromDB.Scopes[0].CredentialType)
		assert.Equal(t, verificationQuery.Scopes[0].CredentialSubject.Bytes, verificationQueryFromDB.Scopes[0].CredentialSubject.Bytes)
		assert.Equal(t, verificationQuery.Scopes[1].ID, verificationQueryFromDB.Scopes[1].ID)
		assert.Equal(t, verificationQuery.Scopes[1].ScopeID, verificationQueryFromDB.Scopes[1].ScopeID)
		assert.Equal(t, verificationQuery.Scopes[1].CircuitID, verificationQueryFromDB.Scopes[1].CircuitID)
		assert.Equal(t, verificationQuery.Scopes[1].Context, verificationQueryFromDB.Scopes[1].Context)
		assert.Equal(t, verificationQuery.Scopes[1].AllowedIssuers, verificationQueryFromDB.Scopes[1].AllowedIssuers)
		assert.Equal(t, verificationQuery.Scopes[1].CredentialType, verificationQueryFromDB.Scopes[1].CredentialType)
		assert.Equal(t, verificationQuery.Scopes[1].CredentialSubject.Bytes, verificationQueryFromDB.Scopes[1].CredentialSubject.Bytes)
	})

	t.Run("should not get the verification", func(t *testing.T) {
		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, uuid.New())
		require.Error(t, err)
		require.Equal(t, VerificationQueryNotFoundError, err)
		assert.Nil(t, verificationQueryFromDB)
	})
}

func TestUpdateVerificationQuery(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:x7tz1NB9fy4GJJW1oQV1wGYpuratuApN8FWEQVKZP"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should update the verification query", func(t *testing.T) {
		credentialSubject1 := pgtype.JSONB{}
		err = credentialSubject1.Set(`{"birthday": {"$eq": 19791109}}`)
		credentialSubject2 := pgtype.JSONB{}
		err = credentialSubject2.Set(`{"position": {"$eq": 1}}`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           1,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
					AllowedIssuers:    []string{"issuer1", "issuer2"},
					CredentialType:    "KYCAgeCredential",
					CredentialSubject: credentialSubject1,
				},
			},
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)

		verificationQuery.Scopes = append(verificationQuery.Scopes, domain.VerificationScope{
			ID:                uuid.New(),
			ScopeID:           2,
			CircuitID:         "credentialAtomicQuerySigV2",
			Context:           "ipfs://QmaBJzpoYT2CViDx5ShJiuYLKXizrPEfXo8JqzrXCvG6oc",
			AllowedIssuers:    []string{"*"},
			CredentialType:    "TestInteger01",
			CredentialSubject: credentialSubject2,
		})

		verificationQuery.SkipCheckRevocation = true
		verificationQuery.ChainID = 137
		_, err = verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)

		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, verificationQueryID)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryFromDB.ID)
		assert.Equal(t, verificationQuery.ChainID, verificationQueryFromDB.ChainID)
		assert.Equal(t, verificationQuery.SkipCheckRevocation, verificationQueryFromDB.SkipCheckRevocation)
		assert.Equal(t, verificationQuery.Scopes[0].ID, verificationQueryFromDB.Scopes[0].ID)
		assert.Equal(t, verificationQuery.Scopes[0].ScopeID, verificationQueryFromDB.Scopes[0].ScopeID)
		assert.Equal(t, verificationQuery.Scopes[0].CircuitID, verificationQueryFromDB.Scopes[0].CircuitID)
		assert.Equal(t, verificationQuery.Scopes[0].Context, verificationQueryFromDB.Scopes[0].Context)
		assert.Equal(t, verificationQuery.Scopes[0].AllowedIssuers, verificationQueryFromDB.Scopes[0].AllowedIssuers)
		assert.Equal(t, verificationQuery.Scopes[0].CredentialType, verificationQueryFromDB.Scopes[0].CredentialType)
		assert.Equal(t, verificationQuery.Scopes[0].CredentialSubject.Bytes, verificationQueryFromDB.Scopes[0].CredentialSubject.Bytes)
		assert.Equal(t, verificationQuery.Scopes[1].ID, verificationQueryFromDB.Scopes[1].ID)
		assert.Equal(t, verificationQuery.Scopes[1].ScopeID, verificationQueryFromDB.Scopes[1].ScopeID)
		assert.Equal(t, verificationQuery.Scopes[1].CircuitID, verificationQueryFromDB.Scopes[1].CircuitID)
		assert.Equal(t, verificationQuery.Scopes[1].Context, verificationQueryFromDB.Scopes[1].Context)
		assert.Equal(t, verificationQuery.Scopes[1].AllowedIssuers, verificationQueryFromDB.Scopes[1].AllowedIssuers)
		assert.Equal(t, verificationQuery.Scopes[1].CredentialType, verificationQueryFromDB.Scopes[1].CredentialType)
		assert.Equal(t, verificationQuery.Scopes[1].CredentialSubject.Bytes, verificationQueryFromDB.Scopes[1].CredentialSubject.Bytes)
	})
}

func TestGetAllVerification(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xCu8Cshrj4oegWRabzGtbKzqUFtXN85x8XkCPdREU"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("GetAll", func(t *testing.T) {
		credentialSubject := pgtype.JSONB{}
		err = credentialSubject.Set(`{"birthday": {"$eq": 19791109}}`)
		credentialSubject2 := pgtype.JSONB{}
		err = credentialSubject2.Set(`{"position": {"$eq": 1}}`)
		require.NoError(t, err)
		verificationQuery1 := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           1,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
					AllowedIssuers:    []string{"issuer1", "issuer2"},
					CredentialType:    "KYCAgeCredential",
					CredentialSubject: credentialSubject,
				},
			},
		}

		verificationQuery2 := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           2,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "ipfs://QmaBJzpoYT2CViDx5ShJiuYLKXizrPEfXo8JqzrXCvG6oc",
					AllowedIssuers:    []string{"*"},
					CredentialType:    "TestInteger01",
					CredentialSubject: credentialSubject2,
				},
			},
		}

		verificationQueryID1, err := verificationRepository.Save(ctx, *did, verificationQuery1)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery1.ID, verificationQueryID1)

		verificationQueryID2, err := verificationRepository.Save(ctx, *did, verificationQuery2)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery2.ID, verificationQueryID2)

		verificationQueryFromDB, err := verificationRepository.GetAll(ctx, *did)
		require.NoError(t, err)
		assert.Equal(t, 2, len(verificationQueryFromDB))
		assert.Equal(t, verificationQuery1.ID, verificationQueryFromDB[0].ID)
		assert.Equal(t, verificationQuery1.ChainID, verificationQueryFromDB[0].ChainID)
		assert.Equal(t, verificationQuery1.SkipCheckRevocation, verificationQueryFromDB[0].SkipCheckRevocation)
		assert.Equal(t, verificationQuery1.Scopes[0].ID, verificationQueryFromDB[0].Scopes[0].ID)
		assert.Equal(t, verificationQuery1.Scopes[0].ScopeID, verificationQueryFromDB[0].Scopes[0].ScopeID)
		assert.Equal(t, verificationQuery1.Scopes[0].CircuitID, verificationQueryFromDB[0].Scopes[0].CircuitID)
		assert.Equal(t, verificationQuery1.Scopes[0].Context, verificationQueryFromDB[0].Scopes[0].Context)
		assert.Equal(t, verificationQuery1.Scopes[0].AllowedIssuers, verificationQueryFromDB[0].Scopes[0].AllowedIssuers)
		assert.Equal(t, verificationQuery1.Scopes[0].CredentialType, verificationQueryFromDB[0].Scopes[0].CredentialType)
		assert.Equal(t, verificationQuery1.Scopes[0].CredentialSubject.Bytes, verificationQueryFromDB[0].Scopes[0].CredentialSubject.Bytes)
		assert.Equal(t, verificationQuery2.ID, verificationQueryFromDB[1].ID)
		assert.Equal(t, verificationQuery2.ChainID, verificationQueryFromDB[1].ChainID)
		assert.Equal(t, verificationQuery2.SkipCheckRevocation, verificationQueryFromDB[1].SkipCheckRevocation)
		assert.Equal(t, verificationQuery2.Scopes[0].ID, verificationQueryFromDB[1].Scopes[0].ID)
		assert.Equal(t, verificationQuery2.Scopes[0].ScopeID, verificationQueryFromDB[1].Scopes[0].ScopeID)
		assert.Equal(t, verificationQuery2.Scopes[0].CircuitID, verificationQueryFromDB[1].Scopes[0].CircuitID)
		assert.Equal(t, verificationQuery2.Scopes[0].Context, verificationQueryFromDB[1].Scopes[0].Context)
		assert.Equal(t, verificationQuery2.Scopes[0].AllowedIssuers, verificationQueryFromDB[1].Scopes[0].AllowedIssuers)
		assert.Equal(t, verificationQuery2.Scopes[0].CredentialType, verificationQueryFromDB[1].Scopes[0].CredentialType)
		assert.Equal(t, verificationQuery2.Scopes[0].CredentialSubject.Bytes, verificationQueryFromDB[1].Scopes[0].CredentialSubject.Bytes)
	})
}

func TestAddVerification(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xCd1tRmXnqbgiT3QC2CuDddUoHK4S9iXwq5xFDJGb"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should add a response to verification", func(t *testing.T) {
		credentialSubject := pgtype.JSONB{}
		err = credentialSubject.Set(`{"birthday": {"$eq": 19791109}}`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scopes: []domain.VerificationScope{
				{
					ID:                uuid.New(),
					ScopeID:           1,
					CircuitID:         "credentialAtomicQuerySigV2",
					Context:           "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld",
					AllowedIssuers:    []string{"issuer1", "issuer2"},
					CredentialType:    "KYCAgeCredential",
					CredentialSubject: credentialSubject,
				},
			},
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)

		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, verificationQueryID)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryFromDB.ID)

		response := pgtype.JSONB{}
		err = response.Set(`{"something": {"proof": 1}}`)
		require.NoError(t, err)
		verificationResponse := domain.VerificationResponse{
			ID:                  uuid.New(),
			VerificationScopeID: verificationQueryFromDB.Scopes[0].ID,
			UserDID:             "did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3",
			Response:            response,
		}

		responseID, err := verificationRepository.AddResponse(ctx, verificationQueryFromDB.Scopes[0].ID, verificationResponse)
		require.NoError(t, err)
		assert.Equal(t, verificationResponse.ID, responseID)
	})

	t.Run("should get an error", func(t *testing.T) {
		response := pgtype.JSONB{}
		err = response.Set(`{"something": {"proof": 1}}`)
		require.NoError(t, err)
		verificationResponse := domain.VerificationResponse{
			ID:       uuid.New(),
			UserDID:  "did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3",
			Response: response,
		}
		responseID, err := verificationRepository.AddResponse(ctx, uuid.New(), verificationResponse)
		require.Error(t, err)
		require.True(t, errors.Is(err, VerificationScopeNotFoundError))
		assert.Equal(t, uuid.Nil, responseID)
	})
}
