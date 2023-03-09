package repositories

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type connections struct{}

// NewConnections returns a new connections repository
func NewConnections() ports.ConnectionsRepository {
	return &connections{}
}

// Save stores in the database the given connection and updates the modified at in case already exists
func (c *connections) Save(ctx context.Context, conn db.Querier, connection *domain.Connection) error {
	sql := `INSERT INTO connections (issuer_id, user_id, issuer_doc, user_doc)
			VALUES($1, $2, $3, $4) ON CONFLICT (issuer_id, user_id) DO
			UPDATE SET issuer_id=$1, user_id=$2, issuer_doc=$3, user_doc=$4,
			           modified_at = now();`
	_, err := conn.Exec(ctx, sql, connection.IssuerDID.String(), connection.UserDID.String(), connection.IssuerDoc, connection.UserDoc)

	return err
}
