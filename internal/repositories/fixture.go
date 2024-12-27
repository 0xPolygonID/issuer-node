package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// Fixture - Handle testing fixture configuration
type Fixture struct {
	storage                 *db.Storage
	identityRepository      ports.IdentityRepository
	claimRepository         ports.ClaimRepository
	connectionsRepository   ports.ConnectionRepository
	schemaRepository        ports.SchemaRepository
	identityStateRepository ports.IdentityStateRepository
	paymentRepository       ports.PaymentRepository
}

// NewFixture - constructor
func NewFixture(storage *db.Storage) *Fixture {
	return &Fixture{
		storage:                 storage,
		identityRepository:      NewIdentity(),
		claimRepository:         NewClaim(),
		connectionsRepository:   NewConnection(),
		schemaRepository:        NewSchema(*storage),
		identityStateRepository: NewIdentityState(),
		paymentRepository:       NewPayment(*storage),
	}
}

// ExecQueryParams - handle the query and the arguments for that query.
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
