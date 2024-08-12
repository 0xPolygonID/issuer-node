package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateIdentity creates a new identity
func (f *Fixture) CreateIdentity(t *testing.T, identity *domain.Identity) {
	t.Helper()
	assert.NoError(t, f.identityRepository.Save(context.Background(), f.storage.Pgx, identity))
}

func (f *Fixture) CreateIdentityStatus(t *testing.T, state domain.IdentityState) {
	t.Helper()
	assert.NoError(t, f.identityStateRepository.Save(context.Background(), f.storage.Pgx, state))
}
