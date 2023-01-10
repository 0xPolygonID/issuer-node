package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

func (f *Fixture) CreateIdentity(t *testing.T, identity *domain.Identity) {
	t.Helper()

	_, err := f.storage.Pgx.Exec(context.Background(), `INSERT INTO identities (identifier, relay, immutable) VALUES ($1, $2, $3)`,
		identity.Identifier, identity.Relay, identity.Immutable)
	assert.NoError(t, err)
}
