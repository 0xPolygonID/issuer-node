package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrSchemaDoesNotExist claim does not exist
var ErrSchemaDoesNotExist = errors.New("schema does not exist")

type dbSchema struct {
	ID         uuid.UUID
	IssuerID   string
	URL        string
	Type       string
	Hash       string
	Attributes string
	CreatedAt  time.Time
}

type schema struct {
	conn db.Storage
}

// NewSchema returns a new schema repository
func NewSchema(conn db.Storage) *schema {
	return &schema{conn: conn}
}

// Save stores a new entry in schemas table
func (r *schema) Save(ctx context.Context, s *domain.Schema) error {
	const insertSchema = `INSERT INTO schemas (id, issuer_id, url, type, attributes, hash, created_at) VALUES($1, $2::text, $3::text, $4::text, $5::text, $6::text, $7);`
	hash, err := s.Hash.MarshalText()
	if err != nil {
		return err
	}
	_, err = r.conn.Pgx.Exec(ctx, insertSchema, s.ID, s.IssuerDID.String(), s.URL, s.Type, s.Attributes.String(), string(hash), s.CreatedAt)
	return err
}

// GetAll returns all the schemas that match the query filter
func (r *schema) GetAll(ctx context.Context, _ *string) ([]domain.Schema, error) {
	const all = `SELECT id, issuer_id, url, type, attributes, hash, created_at FROM schemas ORDER BY created_at DESC`
	rows, err := r.conn.Pgx.Query(ctx, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	schemaCol := make([]domain.Schema, 0)
	s := dbSchema{}
	for rows.Next() {
		if err := rows.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.Attributes, &s.Hash, &s.CreatedAt); err != nil {
			return nil, err
		}
		item, err := toSchemaDomain(&s)
		if err != nil {
			return nil, err
		}
		schemaCol = append(schemaCol, *item)
	}
	return schemaCol, nil
}

// GetByID searches and returns an schema by id
func (r *schema) GetByID(ctx context.Context, id uuid.UUID) (*domain.Schema, error) {
	const byID = `SELECT id, issuer_id, url, type, attributes, hash, created_at 
		FROM schemas 
		WHERE id=$1`

	s := dbSchema{}
	row := r.conn.Pgx.QueryRow(ctx, byID, id)
	err := row.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.Attributes, &s.Hash, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrSchemaDoesNotExist
	}
	if err != nil {
		return nil, err
	}
	return toSchemaDomain(&s)
}

func toSchemaDomain(s *dbSchema) (*domain.Schema, error) {
	issuerDID, err := core.ParseDID(s.IssuerID)
	if err != nil {
		return nil, fmt.Errorf("parsing issuer DID from schema: %w", err)
	}
	schemaHash, err := core.NewSchemaHashFromHex(s.Hash)
	if err != nil {
		return nil, fmt.Errorf("parsing hash from schema: %w", err)
	}
	return &domain.Schema{
		ID:         s.ID,
		IssuerDID:  *issuerDID,
		URL:        s.URL,
		Type:       s.Type,
		Hash:       schemaHash,
		Attributes: domain.SchemaAttrsFromString(s.Attributes),
		CreatedAt:  s.CreatedAt,
	}, nil
}
