package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/db"
)

var queries = []string{
	`
INSERT INTO claims
  (
    id, identifier, issuer, schema_hash, schema_url, schema_type,
    other_identifier, expiration, version, rev_nonce, core_claim,
    identity_state
  )
VALUES
  (
    'bae98b8d-e32e-4990-9e97-1f2fda420914',
    'did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ',
    'did:iden3:polygon:mumbai:wyFiV4w71QgWPn6bYLsZoysFay66gKtVa9kfu6yMZ',
    'ca938857241db9451ea329256b9c06e5',
    'https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld',
    'AuthBJJCredential', '', 0, 0, 0, '["0","0","0","0","0","0","0","0"]',
    NULL
  ),
  (
    '7d78b799-03ea-4b34-8703-b4c677ee0326',
    'did:iden3:polygon:mumbai:x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp',
    'did:iden3:polygon:mumbai:x2Uw18ATvY7mEsgfrrDipBmQQdPWAao4NmF56wGvp',
    'ca938857241db9451ea329256b9c06e5',
    'https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json-ld/auth.json-ld',
    'AuthBJJCredential', '', 0, 0, 0, '["0","0","0","0","0","0","0","0"]',
    '6313885808cc4d87c87bbeff937376646b4a6cce024e6af598cad7cf8d50be25'
  )
`,
	`
INSERT INTO identity_states (state, status)
VALUES
  ('6313885808cc4d87c87bbeff937376646b4a6cce024e6af598cad7cf8d50be25', 'confirmed')
`,
}

type Fixture struct {
	storage *db.Storage
}

func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage: storage,
	}
}

func (f *Fixture) execQuery(t *testing.T, query string) {
	t.Helper()
	_, err := f.storage.Pgx.Exec(context.Background(), query)
	assert.NoError(t, err)
}
