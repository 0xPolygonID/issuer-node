package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrConnectionDoesNotExist connection does not exist
var ErrConnectionDoesNotExist = errors.New("connection does not exist")

type connections struct{}

// NewConnections returns a new connections repository
func NewConnections() ports.ConnectionsRepository {
	return &connections{}
}

// Save stores in the database the given connection and updates the modified at in case already exists
func (c *connections) Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error) {
	var id uuid.UUID
	sql := `INSERT INTO connections (id,issuer_id, user_id, issuer_doc, user_doc,created_at,modified_at)
			VALUES($1, $2, $3, $4,$5,$6,$7) ON CONFLICT (issuer_id, user_id) DO
			UPDATE SET issuer_id=$2, user_id=$3, issuer_doc=$4, user_doc=$5,
			           modified_at = $7
	RETURNING id;`
	err := conn.QueryRow(ctx, sql, connection.ID, connection.IssuerDID.String(), connection.UserDID.String(), connection.IssuerDoc, connection.UserDoc, connection.CreatedAt, connection.ModifiedAt).Scan(&id)

	return id, err
}

func (c *connections) Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error {
	sql := `DELETE FROM connections WHERE id = $1`
	cmd, err := conn.Exec(ctx, sql, id.String())
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrConnectionDoesNotExist
	}

	return nil
}
