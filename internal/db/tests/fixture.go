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

type ExecQueryParams struct {
	Query     string
	Arguments []interface{}
}

func (f *Fixture) ExecQuery(t *testing.T, params ExecQueryParams) {
	t.Helper()
	_, err := f.storage.Pgx.Exec(context.Background(), params.Query, params.Arguments...)
	assert.NoError(t, err)
}
