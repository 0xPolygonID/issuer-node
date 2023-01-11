package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

type Fixture struct {
	storage            *db.Storage
	identityRepository ports.IndentityRepository
	claimRepository    ports.ClaimsRepository
}

func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage:            storage,
		identityRepository: repositories.NewIdentity(),
		claimRepository:    repositories.NewClaims(),
	}
}

func (f *Fixture) execQuery(t *testing.T, query string) {
	t.Helper()
	_, err := f.storage.Pgx.Exec(context.Background(), query)
	assert.NoError(t, err)
}
