package services

import (
	b64 "encoding/base64"
	"testing"
	"time"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

func TestCreateKey(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)

	t.Run("should create a key and auth core claim", func(t *testing.T) {
		keyID, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key")
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		decodedKeyID, err := b64.StdEncoding.DecodeString(keyID.ID)
		assert.NoError(t, err)

		revNonce := common.ToPointer[uint64](uint64(10))
		expiration := common.ToPointer(time.Now().Add(365 * 24 * time.Hour))
		version := common.ToPointer(uint32(1))
		authCoreClaimID, err := identityService.CreateAuthCredential(ctx, did, string(decodedKeyID), revNonce, expiration, version, verifiable.Iden3commRevocationStatusV1)
		assert.NoError(t, err)
		assert.NotNil(t, authCoreClaimID)
		kmsKey, err := keyService.Get(ctx, did, string(decodedKeyID))
		assert.NoError(t, err)
		assert.NotNil(t, kmsKey)
		assert.Equal(t, "test key", kmsKey.Name)
		assert.Equal(t, kms.KeyTypeBabyJubJub, kmsKey.KeyType)
		assert.Equal(t, string(decodedKeyID), kmsKey.KeyID)
		assert.NotEmpty(t, kmsKey.PublicKey)
		assert.True(t, kmsKey.HasAssociatedAuthCredential)
	})
}

func TestGetKey(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)

	t.Run("should get key", func(t *testing.T) {
		keyID, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key")
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		decodedKeyID, err := b64.StdEncoding.DecodeString(keyID.ID)
		assert.NoError(t, err)
		kmsKey, err := keyService.Get(ctx, did, string(decodedKeyID))
		assert.NoError(t, err)
		assert.NotNil(t, kmsKey)
		assert.Equal(t, "test key", kmsKey.Name)
		assert.Equal(t, kms.KeyTypeBabyJubJub, kmsKey.KeyType)
		assert.Equal(t, string(decodedKeyID), kmsKey.KeyID)
		assert.NotEmpty(t, kmsKey.PublicKey)
		assert.False(t, kmsKey.HasAssociatedAuthCredential)
	})

	t.Run("should return an error", func(t *testing.T) {
		kmsKey, err := keyService.Get(ctx, did, "123")
		assert.Error(t, err)
		assert.Nil(t, kmsKey)
	})
}

func TestGetAll(t *testing.T) {
	ctx := t.Context()
	identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
		Blockchain: blockchain,
		Network:    net,
		Method:     method,
	})
	assert.NoError(t, err)
	assert.NotNil(t, identity)
	identifier := identity.Identifier
	did, err := w3c.ParseDID(identifier)
	assert.NoError(t, err)
	keyID1, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key")
	assert.NoError(t, err)
	assert.NotNil(t, keyID1)
	keyID2, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key 2")
	assert.NoError(t, err)
	assert.NotNil(t, keyID2)

	keys, total, err := keyService.GetAll(ctx, did, ports.KeyFilter{MaxResults: 10, Page: 1})
	assert.NoError(t, err)
	assert.Equal(t, uint(3), total)
	assert.Len(t, keys, 3)

	keys, total, err = keyService.GetAll(ctx, did, ports.KeyFilter{MaxResults: 2, Page: 1})
	assert.NoError(t, err)
	assert.Equal(t, uint(3), total)
	assert.Len(t, keys, 2)

	keys, total, err = keyService.GetAll(ctx, did, ports.KeyFilter{MaxResults: 2, Page: 2})
	assert.NoError(t, err)
	assert.Equal(t, uint(3), total)
	assert.Len(t, keys, 1)
}

func TestDeleteKey(t *testing.T) {
	ctx := t.Context()
	t.Run("should delete key", func(t *testing.T) {
		identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
			Blockchain: blockchain,
			Network:    net,
			Method:     method,
		})
		assert.NoError(t, err)
		assert.NotNil(t, identity)
		identifier := identity.Identifier
		did, err := w3c.ParseDID(identifier)
		assert.NoError(t, err)

		keyID, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key")
		assert.NoError(t, err)
		assert.NotNil(t, keyID)

		decodedKeyID, err := b64.StdEncoding.DecodeString(keyID.ID)
		assert.NoError(t, err)
		assert.NoError(t, keyService.Delete(ctx, did, string(decodedKeyID)))
	})

	t.Run("should not delete key", func(t *testing.T) {
		identity, err := identityService.Create(ctx, "http://localhost", &ports.DIDCreationOptions{
			Blockchain: blockchain,
			Network:    net,
			Method:     method,
		})
		assert.NoError(t, err)
		assert.NotNil(t, identity)
		identifier := identity.Identifier
		did, err := w3c.ParseDID(identifier)
		assert.NoError(t, err)

		keyID, err := keyService.Create(ctx, did, kms.KeyTypeBabyJubJub, "test key")
		assert.NoError(t, err)
		assert.NotNil(t, keyID)
		decodedKeyID, err := b64.StdEncoding.DecodeString(keyID.ID)
		assert.NoError(t, err)
		revNonce := common.ToPointer[uint64](uint64(10))
		expiration := common.ToPointer(time.Now().Add(365 * 24 * time.Hour))
		version := common.ToPointer(uint32(1))
		authCoreClaimID, err := identityService.CreateAuthCredential(ctx, did, string(decodedKeyID), revNonce, expiration, version, verifiable.Iden3commRevocationStatusV1)
		assert.NoError(t, err)
		assert.NotNil(t, authCoreClaimID)
		assert.Error(t, keyService.Delete(ctx, did, string(decodedKeyID)))
	})
}
