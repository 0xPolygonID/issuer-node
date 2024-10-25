package kms

import (
	"fmt"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"
)

// getKeyID returns key ID string
// if identity is nil, key ID is returned as is keyType:keyID (BJJ:PrivateKey)
// if identity is not nil and keyID contains keyType, key ID is returned as identity/keyType:keyID (did/BJJ:PrivateKey)
// if identity is not nil and keyID does not contain keyType, key ID is returned as identity/keyID (did:PrivateKey)
func getKeyID(identity *w3c.DID, keyType KeyType, keyID string) string {
	if identity == nil {
		return fmt.Sprintf("%v:%v", keyType, keyID)
	} else {
		if !strings.Contains(keyID, string(keyType)) {
			return fmt.Sprintf("%v/%v:%v", identity.String(), keyType, keyID)
		}
		return fmt.Sprintf("%v/%v", identity.String(), keyID)
	}
}

// convertToKeyType - converts string to KeyType
// converts from babyjubjub to BJJ and from ethereum to ETH
func convertToKeyType(keyType string) KeyType {
	switch keyType {
	case babyjubjub:
		return KeyTypeBabyJubJub
	case ethereum:
		return KeyTypeEthereum
	default:
		return ""
	}
}

// convertFromKeyType - converts KeyType to string
// converts from BJJ to babyjubjub and from ETH to ethereum
func convertFromKeyType(keyType KeyType) string {
	switch keyType {
	case KeyTypeBabyJubJub:
		return babyjubjub
	case KeyTypeEthereum:
		return ethereum
	default:
		return ""
	}
}
