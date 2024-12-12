package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgconn"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/log"
)

var (
	// ErrKeyNotFound key not found error
	ErrKeyNotFound = errors.New("key not found")
	// ErrDuplicateKeyName duplicate key name error
	ErrDuplicateKeyName = errors.New("key name already exists")
)

type key struct {
	conn db.Storage
}

// NewKey returns a new key repository
func NewKey(conn db.Storage) *key {
	return &key{
		conn,
	}
}

// Save saves a key
func (k *key) Save(ctx context.Context, conn db.Querier, key *domain.Key) (uuid.UUID, error) {
	if conn == nil {
		conn = k.conn.Pgx
	}
	sql := `INSERT INTO keys (id, issuer_did, public_key, name)
			VALUES($1, $2, $3, $4) ON CONFLICT (id) DO
			UPDATE SET name=$4`
	_, err := conn.Exec(ctx, sql, key.ID, key.IssuerCoreDID().String(), key.PublicKey, key.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateViolationErrorCode {
			return uuid.Nil, ErrDuplicateKeyName
		}
		return uuid.Nil, err
	}
	return key.ID, err
}

// GetByPublicKey returns a key by its public key
func (k *key) GetByPublicKey(ctx context.Context, issuerDID w3c.DID, publicKey string) (*domain.Key, error) {
	sql := `SELECT id, issuer_did, public_key, name 
			FROM keys WHERE  issuer_did=$1 and public_key=$2`
	row := k.conn.Pgx.QueryRow(ctx, sql, issuerDID.String(), publicKey)

	key := domain.Key{}
	err := row.Scan(&key.ID, &key.IssuerDID, &key.PublicKey, &key.Name)
	if err != nil {
		log.Error(ctx, "error getting key by public key", "err", err)
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return &key, nil
}

// Delete deletes a key by its public key
func (k *key) Delete(ctx context.Context, issuerDID w3c.DID, publicKey string) error {
	sql := `DELETE FROM keys WHERE issuer_did=$1 AND public_key=$2`
	_, err := k.conn.Pgx.Exec(ctx, sql, issuerDID.String(), publicKey)
	return err
}
