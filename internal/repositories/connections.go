package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"

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
	sql := `INSERT INTO connections (id,issuer_id, user_id, issuer_doc, user_doc,created_at,modified_at)
			VALUES($1, $2, $3, $4,$5,$6,$7) ON CONFLICT (issuer_id, user_id) DO
			UPDATE SET issuer_id=$2, user_id=$3, issuer_doc=$4, user_doc=$5,
			           modified_at = now();`
	_, err := conn.Exec(ctx, sql, uuid.New().String(), connection.IssuerDID.String(), connection.UserDID.String(), connection.IssuerDoc, connection.UserDoc, time.Now(), time.Now())

	return err
}
