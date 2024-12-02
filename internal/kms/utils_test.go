package kms

import (
	"testing"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
)

func Test_getKeyID(t *testing.T) {
	identity, err := w3c.ParseDID("did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3")
	assert.NoError(t, err)

	t.Run("should get did/BJJ:PrivateKey id", func(t *testing.T) {
		keyType := KeyTypeBabyJubJub
		id := getKeyID(identity, keyType, "PrivateKey")
		assert.Equal(t, "did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3/BJJ:PrivateKey", id)
	})

	t.Run("should get did/ETH:PrivateKey id", func(t *testing.T) {
		keyType := KeyTypeEthereum
		id := getKeyID(identity, keyType, "PrivateKey")
		assert.Equal(t, "did:iden3:privado:main:2SizDYDWBViKXRfp1VgUAMqhz5SDvP7D1MYiPfwJV3/ETH:PrivateKey", id)
	})

	t.Run("should get ETH:PrivateKey id", func(t *testing.T) {
		keyType := KeyTypeEthereum
		id := getKeyID(nil, keyType, "PrivateKey")
		assert.Equal(t, "ETH:PrivateKey", id)
	})

	t.Run("should get BJJ:PrivateKey id", func(t *testing.T) {
		keyType := KeyTypeBabyJubJub
		id := getKeyID(nil, keyType, "PrivateKey")
		assert.Equal(t, "BJJ:PrivateKey", id)
	})
}
