package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/issuer-node/internal/core/ports"
	"github.com/polygonid/issuer-node/internal/db"
	"github.com/polygonid/issuer-node/internal/repositories"
)

// Fixture - Handle testing fixture configuration
type Fixture struct {
	storage               *db.Storage
	identityRepository    ports.IndentityRepository
	claimRepository       ports.ClaimsRepository
	connectionsRepository ports.ConnectionsRepository
	schemaRepository      ports.SchemaRepository
}

// NewFixture - constructor
func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage:               storage,
		identityRepository:    repositories.NewIdentity(),
		claimRepository:       repositories.NewClaims(),
		connectionsRepository: repositories.NewConnections(),
		schemaRepository:      repositories.NewSchema(*storage),
	}
}

// ExecQueryParams - handle the query and the argumens for that query.
type ExecQueryParams struct {
	Query     string
	Arguments []interface{}
}

// ExecQuery - Execute a query for testing purpose.
func (f *Fixture) ExecQuery(t *testing.T, params ExecQueryParams) {
	t.Helper()
	_, err := f.storage.Pgx.Exec(context.Background(), params.Query, params.Arguments...)
	assert.NoError(t, err)
}
