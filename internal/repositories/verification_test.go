package repositories

import (
	"context"
	"encoding/json"
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
		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope,
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)
	})
}

//nolint:all
func TestGetVerification(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xBdqiqz3yVT79NEAuNaqKSDZ6a5V6q8Ph66i5d2tT"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should get the verification", func(t *testing.T) {
		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope,
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)

		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, verificationQueryID)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryFromDB.ID)
		assert.Equal(t, verificationQuery.ChainID, verificationQueryFromDB.ChainID)
		assert.Equal(t, verificationQuery.SkipCheckRevocation, verificationQueryFromDB.SkipCheckRevocation)
		assert.NotNil(t, verificationQueryFromDB.Scope)

		var res []map[string]interface{}
		require.NoError(t, json.Unmarshal(verificationQuery.Scope.Bytes, &res))
		assert.Equal(t, 1, len(res))
		assert.Equal(t, 1, int(res[0]["ID"].(float64)))
		assert.Equal(t, "credentialAtomicQuerySigV2", res[0]["circuitID"])
		assert.Equal(t, "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", res[0]["query"].(map[string]interface{})["context"])
		assert.Equal(t, []interface{}{"*"}, res[0]["query"].(map[string]interface{})["allowedIssuers"])
		assert.Equal(t, "KYCAgeCredential", res[0]["query"].(map[string]interface{})["type"])
		assert.Equal(t, 19791109, int(res[0]["query"].(map[string]interface{})["credentialSubject"].(map[string]interface{})["birthday"].(map[string]interface{})["$eq"].(float64)))
	})

	t.Run("should not get the verification", func(t *testing.T) {
		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, uuid.New())
		require.Error(t, err)
		require.Equal(t, ErrVerificationQueryNotFound, err)
		assert.Nil(t, verificationQueryFromDB)
	})
}

//nolint:all
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

		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		scope2 := pgtype.JSONB{}
		err = scope2.Set(`[{"ID": 1,"circuitID": "credentialAtomicQueryV3-beta.1","query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld","allowedIssuers": ["*"],"type": "KYCAgeCredential","credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope,
		}

		verificationQueryID, err := verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryID)

		verificationQuery.Scope = &scope2
		verificationQuery.SkipCheckRevocation = true
		verificationQuery.ChainID = 137
		_, err = verificationRepository.Save(ctx, *did, verificationQuery)
		require.NoError(t, err)

		verificationQueryFromDB, err := verificationRepository.Get(ctx, *did, verificationQueryID)
		require.NoError(t, err)
		assert.Equal(t, verificationQuery.ID, verificationQueryFromDB.ID)
		assert.Equal(t, verificationQuery.ChainID, verificationQueryFromDB.ChainID)
		assert.Equal(t, verificationQuery.SkipCheckRevocation, verificationQueryFromDB.SkipCheckRevocation)
		assert.NotNil(t, verificationQueryFromDB.Scope)

		var res []map[string]interface{}
		require.NoError(t, json.Unmarshal(verificationQueryFromDB.Scope.Bytes, &res))
		assert.Equal(t, 1, len(res))
		assert.Equal(t, 1, int(res[0]["ID"].(float64)))
		assert.Equal(t, "credentialAtomicQueryV3-beta.1", res[0]["circuitID"])
		assert.Equal(t, "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", res[0]["query"].(map[string]interface{})["context"])
		assert.Equal(t, []interface{}{"*"}, res[0]["query"].(map[string]interface{})["allowedIssuers"])
		assert.Equal(t, "KYCAgeCredential", res[0]["query"].(map[string]interface{})["type"])
		assert.Equal(t, 19791109, int(res[0]["query"].(map[string]interface{})["credentialSubject"].(map[string]interface{})["birthday"].(map[string]interface{})["$eq"].(float64)))
	})
}

//nolint:all
func TestGetAllVerification(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xCu8Cshrj4oegWRabzGtbKzqUFtXN85x8XkCPdREU"
	verificationRepository := NewVerification(*storage)

	_, err := storage.Pgx.Exec(ctx, "INSERT INTO identities (identifier, keytype) VALUES ($1, $2)", didStr, "BJJ")
	assert.NoError(t, err)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("GetAll", func(t *testing.T) {
		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)

		scope2 := pgtype.JSONB{}
		err = scope2.Set(`[{"ID": 2,"circuitID": "credentialAtomicQueryV3-beta.1","query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld","allowedIssuers": ["*"],"type": "KYCAgeCredential","credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)

		verificationQuery1 := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope,
		}

		verificationQuery2 := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope2,
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

		var resVerificationQuery1 []map[string]interface{}
		require.NoError(t, json.Unmarshal(verificationQueryFromDB[0].Scope.Bytes, &resVerificationQuery1))

		assert.Equal(t, verificationQuery1.ID, verificationQueryFromDB[0].ID)
		assert.Equal(t, verificationQuery1.ChainID, verificationQueryFromDB[0].ChainID)
		assert.Equal(t, verificationQuery1.SkipCheckRevocation, verificationQueryFromDB[0].SkipCheckRevocation)
		assert.Equal(t, 1, len(resVerificationQuery1))
		assert.Equal(t, 1, int(resVerificationQuery1[0]["ID"].(float64)))
		assert.Equal(t, "credentialAtomicQuerySigV2", resVerificationQuery1[0]["circuitID"])
		assert.Equal(t, "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", resVerificationQuery1[0]["query"].(map[string]interface{})["context"])
		assert.Equal(t, []interface{}{"*"}, resVerificationQuery1[0]["query"].(map[string]interface{})["allowedIssuers"])
		assert.Equal(t, "KYCAgeCredential", resVerificationQuery1[0]["query"].(map[string]interface{})["type"])
		assert.Equal(t, 19791109, int(resVerificationQuery1[0]["query"].(map[string]interface{})["credentialSubject"].(map[string]interface{})["birthday"].(map[string]interface{})["$eq"].(float64)))

		var resVerificationQuery2 []map[string]interface{}
		require.NoError(t, json.Unmarshal(verificationQueryFromDB[1].Scope.Bytes, &resVerificationQuery2))

		assert.Equal(t, verificationQuery2.ID, verificationQueryFromDB[1].ID)
		assert.Equal(t, verificationQuery2.ChainID, verificationQueryFromDB[1].ChainID)
		assert.Equal(t, verificationQuery2.SkipCheckRevocation, verificationQueryFromDB[1].SkipCheckRevocation)
		assert.Equal(t, 1, len(resVerificationQuery2))
		assert.Equal(t, 2, int(resVerificationQuery2[0]["ID"].(float64)))
		assert.Equal(t, "credentialAtomicQueryV3-beta.1", resVerificationQuery2[0]["circuitID"])
		assert.Equal(t, "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", resVerificationQuery2[0]["query"].(map[string]interface{})["context"])
		assert.Equal(t, []interface{}{"*"}, resVerificationQuery2[0]["query"].(map[string]interface{})["allowedIssuers"])
		assert.Equal(t, "KYCAgeCredential", resVerificationQuery2[0]["query"].(map[string]interface{})["type"])
		assert.Equal(t, 19791109, int(resVerificationQuery2[0]["query"].(map[string]interface{})["credentialSubject"].(map[string]interface{})["birthday"].(map[string]interface{})["$eq"].(float64)))
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

		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               &scope,
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
			VerificationQueryID: verificationQueryFromDB.ID,
			UserDID:             "did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3",
			Response:            &response,
		}

		responseID, err := verificationRepository.AddResponse(ctx, verificationQueryFromDB.ID, verificationResponse)
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
			Response: &response,
		}
		responseID, err := verificationRepository.AddResponse(ctx, uuid.New(), verificationResponse)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrVerificationQueryNotFound))
		assert.Equal(t, uuid.Nil, responseID)
	})
}

func TestGetVerificationResponse(t *testing.T) {
	ctx := context.Background()
	didStr := "did:iden3:polygon:amoy:xCd1tRmXnqbgiT3QC2CuDddUoHK4S9iXwq5xFDJGb"
	verificationRepository := NewVerification(*storage)

	did, err := w3c.ParseDID(didStr)
	require.NoError(t, err)

	t.Run("should get a response to verification", func(t *testing.T) {
		credentialSubject := pgtype.JSONB{}
		err = credentialSubject.Set(`{"birthday": {"$eq": 19791109}}`)
		require.NoError(t, err)

		scope := pgtype.JSONB{}
		err = scope.Set(`[{"ID": 1, "circuitID": "credentialAtomicQuerySigV2", "query": {"context": "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/kyc-v3.json-ld", "allowedIssuers": ["*"], "type": "KYCAgeCredential", "credentialSubject": {"birthday": {"$eq": 19791109}}}}]`)
		require.NoError(t, err)
		verificationQuery := domain.VerificationQuery{
			ID:                  uuid.New(),
			ChainID:             8002,
			SkipCheckRevocation: false,
			Scope:               nil,
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
			VerificationQueryID: verificationQueryFromDB.ID,
			UserDID:             "did:iden3:privado:main:2SizDYDWBVi",
			Response:            &response,
		}

		responseID, err := verificationRepository.AddResponse(ctx, verificationQueryFromDB.ID, verificationResponse)
		require.NoError(t, err)
		assert.Equal(t, verificationResponse.ID, responseID)

		verificationResponseFromDB, err := verificationRepository.GetVerificationResponse(ctx, verificationQueryFromDB.ID)
		require.NoError(t, err)
		assert.Equal(t, verificationResponse.ID, verificationResponseFromDB.ID)
		assert.Equal(t, verificationResponse.UserDID, verificationResponseFromDB.UserDID)
		assert.NotNil(t, verificationResponseFromDB.Response)
	})
}
