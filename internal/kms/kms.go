package kms

import (
	"context"
	stderr "errors"
	"fmt"
	"sync"

	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/pkg/errors"
)

// KMSType represents the KMS interface
// revive:disable-next-line
type KMSType interface {
	RegisterKeyProvider(kt KeyType, kp KeyProvider) error
	CreateKey(kt KeyType, identity *w3c.DID) (KeyID, error)
	PublicKey(keyID KeyID) ([]byte, error)
	Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error)
	KeysByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error)
	LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error)
}

// KeyProvider describes the interface that key providers should match.
type KeyProvider interface {
	// New generates random key.
	// If identity is nil, create new key without binding to identity.
	New(identity *w3c.DID) (KeyID, error)
	// PublicKey returns byte representation of public key
	PublicKey(keyID KeyID) ([]byte, error)
	// Sign the data and return signature.
	Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error)
	// ListByIdentity returns all keys associated with identity
	ListByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error)
	// LinkToIdentity links unbound key to identity.
	// KeyID can be changed after linking.
	// Returning new KeyID.
	LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error)
}

// KMS stores keys and secrets
type KMS struct {
	registry map[KeyType]KeyProvider
}

// KeyType describes the type of Key
type KeyType string

// List of supported key types
const (
	KeyTypeBabyJubJub KeyType = "BJJ"
	KeyTypeEthereum   KeyType = "ETH"
)

// ErrUnknownKeyType returns when we do not support this type of keys
var ErrUnknownKeyType = stderr.New("unknown key type")

// ErrIncorrectKeyType returns when key provider can't work with given key type
var ErrIncorrectKeyType = stderr.New("incorrect key type")

// ErrKeyTypeConflict raises when we register new key provider with key type
// that already exists
var ErrKeyTypeConflict = stderr.New("key type already registered")

// KeyID is a key unique identifier
type KeyID struct {
	Type KeyType
	ID   string
}

// NewKMS create new KMS
func NewKMS() *KMS {
	k := &KMS{registry: make(map[KeyType]KeyProvider)}
	return k
}

// RegisterKeyProvider register new key provider. It is thread unsafe
// function should be called on app initialization or under external mutex.
func (k *KMS) RegisterKeyProvider(kt KeyType, kp KeyProvider) error {
	if _, ok := k.registry[kt]; ok {
		return errors.WithStack(ErrKeyTypeConflict)
	}

	k.registry[kt] = kp
	return nil
}

// CreateKey creates new random key of specified type.
// If identity is not nil, store key for that identity. If nil, do not bind
// key to identity.
func (k *KMS) CreateKey(kt KeyType, identity *w3c.DID) (KeyID, error) {
	var id KeyID
	kp, ok := k.registry[kt]
	if !ok {
		return id, errors.WithStack(ErrUnknownKeyType)
	}
	return kp.New(identity)
}

// PublicKey returns bytes representation for public key for specified key ID
func (k *KMS) PublicKey(keyID KeyID) ([]byte, error) {
	kp, ok := k.registry[keyID.Type]
	if !ok {
		return nil, errors.WithStack(ErrUnknownKeyType)
	}
	return kp.PublicKey(keyID)
}

// Sign signs digest with private key
func (k *KMS) Sign(ctx context.Context, keyID KeyID, data []byte) ([]byte, error) {
	kp, ok := k.registry[keyID.Type]
	if !ok {
		return nil, errors.WithStack(ErrUnknownKeyType)
	}

	return kp.Sign(ctx, keyID, data)
}

// KeysByIdentity lists keys by identity
func (k *KMS) KeysByIdentity(ctx context.Context, identity w3c.DID) ([]KeyID, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	type resT struct {
		keys []KeyID
		err  error
	}

	ch1 := make(chan resT)
	for _, kp := range k.registry {
		wg.Add(1)
		go func(kp KeyProvider) {
			defer wg.Done()
			var r resT
			r.keys, r.err = kp.ListByIdentity(ctx, identity)
			select {
			case ch1 <- r:
			case <-ctx.Done():
			}
		}(kp)
	}

	go func() {
		wg.Wait()
		close(ch1)
	}()

	var allKeys []KeyID
	var err error
	for r := range ch1 {
		if r.err != nil {
			err = r.err
			cancel()
			break
		}
		allKeys = append(allKeys, r.keys...)
	}
	return allKeys, err
}

// LinkToIdentity links unbound key to identity.
// KeyID can be changed after linking.
// Returning new KeyID.
// Old key may be removed after vault. Not all key providers can support this
// operation.
func (k *KMS) LinkToIdentity(ctx context.Context, keyID KeyID, identity w3c.DID) (KeyID, error) {
	kp, ok := k.registry[keyID.Type]
	if !ok {
		return keyID, errors.WithStack(ErrUnknownKeyType)
	}

	return kp.LinkToIdentity(ctx, keyID, identity)
}

// Open returns an initialized KMS
func Open(pluginIden3MountPath string, vault *api.Client) (*KMS, error) {
	bjjKeyProvider, err := NewVaultPluginIden3KeyProvider(vault, pluginIden3MountPath, KeyTypeBabyJubJub)
	if err != nil {
		return nil, fmt.Errorf("cannot create BabyJubJub key provider: %+v", err)
	}

	ethKeyProvider, err := NewVaultPluginIden3KeyProvider(vault, pluginIden3MountPath, KeyTypeEthereum)
	if err != nil {
		return nil, fmt.Errorf("cannot create Ethereum key provider: %+v", err)
	}

	keyStore := NewKMS()
	err = keyStore.RegisterKeyProvider(KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		return nil, fmt.Errorf("cannot register BabyJubJub key provider: %+v", err)
	}

	err = keyStore.RegisterKeyProvider(KeyTypeEthereum, ethKeyProvider)
	if err != nil {
		return nil, fmt.Errorf("cannot register Ethereum key provider: %+v", err)
	}

	return keyStore, nil
}
