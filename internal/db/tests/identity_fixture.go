package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/issuer-node/internal/core/domain"
)

// CreateIdentity creates a new identity
func (f *Fixture) CreateIdentity(t *testing.T, identity *domain.Identity) {
	t.Helper()
	assert.NoError(t, f.identityRepository.Save(context.Background(), f.storage.Pgx, identity))
}
