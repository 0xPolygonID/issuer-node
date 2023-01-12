package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CreateClaim fixture
func (f *Fixture) CreateClaim(t *testing.T, claim *domain.Claim) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id, err := f.claimRepository.Save(ctx, f.storage.Pgx, claim)
	assert.NoError(t, err)
	return id
}
