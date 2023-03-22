package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrConnectionDoesNotExist connection does not exist
var ErrConnectionDoesNotExist = errors.New("connection does not exist")

type dbConnection struct {
	ID         uuid.UUID
	IssuerDID  string
	UserDID    string
	IssuerDoc  pgtype.JSONB
	UserDoc    pgtype.JSONB
	CreatedAt  time.Time
	ModifiedAt time.Time
}

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
	RETURNING id`
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

func (c *connections) GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID core.DID) (*domain.Connection, error) {
	connection := dbConnection{}
	err := conn.QueryRow(ctx,
		`SELECT id, issuer_id,user_id,issuer_doc,user_doc,created_at,modified_at 
				FROM connections 
				WHERE connections.id = $1 AND connections.issuer_id = $2`, id.String(), issuerID.String()).Scan(
		&connection.ID,
		&connection.IssuerDID,
		&connection.UserDID,
		&connection.IssuerDoc,
		&connection.UserDoc,
		&connection.CreatedAt,
		&connection.ModifiedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return toConnectionDomain(&connection)
}

func (c *connections) GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerDID core.DID, query *string) ([]*domain.Connection, error) {
	all := `SELECT id, issuer_id,user_id,issuer_doc,user_doc,created_at,modified_at 
FROM connections 
WHERE connections.issuer_id = $1`
	var err error
	var rows pgx.Rows
	attrs := []interface{}{issuerDID.String()}
	if query != nil && *query != "" {
		did := getDIDFromQuery(*query)
		if did != "" {
			all += ` AND connections.user_id LIKE CONCAT($2::text,'%%')`
			attrs = append(attrs, did)
		}
	}

	rows, err = conn.Query(ctx, all, attrs...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	domainConns := make([]*domain.Connection, 0)
	dbConn := dbConnection{}
	for rows.Next() {
		if err := rows.Scan(&dbConn.ID, &dbConn.IssuerDID, &dbConn.UserDID, &dbConn.IssuerDoc, &dbConn.UserDoc, &dbConn.CreatedAt, &dbConn.ModifiedAt); err != nil {
			return nil, err
		}
		domainConn, err := toConnectionDomain(&dbConn)
		if err != nil {
			return nil, err
		}
		domainConns = append(domainConns, domainConn)
	}

	return domainConns, nil
}

func getDIDFromQuery(query string) string {
	words := strings.Split(strings.ReplaceAll(query, ",", " "), " ")
	for _, word := range words {
		if strings.HasPrefix(word, "did:") {
			return word
		}
	}

	return ""
}

func toConnectionDomain(c *dbConnection) (*domain.Connection, error) {
	issID, err := core.ParseDID(c.IssuerDID)
	if err != nil {
		return nil, fmt.Errorf("parsing issuer DID from connection: %w", err)
	}

	usrDID, err := core.ParseDID(c.UserDID)
	if err != nil {
		return nil, fmt.Errorf("parsing user DID from connection: %w", err)
	}

	conn := &domain.Connection{
		ID:         c.ID,
		IssuerDID:  *issID,
		UserDID:    *usrDID,
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}

	if err := c.UserDoc.AssignTo(&conn.UserDoc); err != nil {
		return nil, fmt.Errorf("parsing user UserDoc from connection: %w", err)
	}

	if err := c.IssuerDoc.AssignTo(&conn.IssuerDoc); err != nil {
		return nil, fmt.Errorf("parsing user IssuerDoc from connection: %w", err)
	}

	return conn, nil
}
