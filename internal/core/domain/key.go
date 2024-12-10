package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// KeyCoreDID is a DID type for keys
type KeyCoreDID w3c.DID

// Key is a key domain model
type Key struct {
	ID        uuid.UUID  `json:"id"`
	IssuerDID KeyCoreDID `json:"issuer_did"`
	PublicKey string     `json:"public_key"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
}

// NewKey creates a new Key
func NewKey(issuerDID w3c.DID, publicKey, name string) *Key {
	return &Key{
		ID:        uuid.New(),
		IssuerDID: KeyCoreDID(issuerDID),
		PublicKey: publicKey,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// IssuerCoreDID returns the issuer DID as a w3c.DID pointer
func (key *Key) IssuerCoreDID() *w3c.DID {
	return common.ToPointer(w3c.DID(key.IssuerDID))
}

// Scan implements the sql.Scanner interface
func (keydid *KeyCoreDID) Scan(value interface{}) error {
	didStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type, expected string")
	}
	did, err := w3c.ParseDID(didStr)
	if err != nil {
		return err
	}
	*keydid = KeyCoreDID(*did)
	return nil
}
