package tests

import (
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// Fixture struct
type Fixture struct {
	storage            *db.Storage
	identityRepository ports.IndentityRepository
	claimRepository    ports.ClaimsRepository
}

// NewFixture returns a new Fixture
func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage:            storage,
		identityRepository: repositories.NewIdentity(),
		claimRepository:    repositories.NewClaims(),
	}
}

//func (f *Fixture) execQuery(t *testing.T, query string) {
//	t.Helper()
//	_, err := f.storage.Pgx.Exec(context.Background(), query)
//	assert.NoError(t, err)
//}
