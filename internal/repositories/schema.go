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
	conn db.Querier
}

// NewSchema returns a new schema repository
func NewSchema(conn db.Querier) *schema {
	return &schema{conn: conn}
}

func (r *schema) Save(ctx context.Context, s *domain.Schema) error {
	const insertSchema = `INSERT INTO schemas (id, issuer_id, url, type, attributes, hash, created_at) VALUES($1, $2::text, $3::text, $4::text, $5::text, $6::text, $7);`
	hash, err := s.Hash.MarshalText()
	if err != nil {
		return err
	}
	_, err = r.conn.Exec(ctx, insertSchema, s.ID, s.IssuerDID.String(), s.URL, s.Type, s.Attributes.String(), string(hash), s.CreatedAt)
	return err
}

func (r *schema) GetById(ctx context.Context, id uuid.UUID) (*domain.Schema, error) {
	const byID = `SELECT id, issuer_id, url, type, attributes, hash, created_at 
		FROM schemas 
		WHERE id=$1`

	s := dbSchema{}
	row := r.conn.QueryRow(ctx, byID, id)
	err := row.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.Attributes, &s.Hash, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
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
		return nil, fmt.Errorf("parsting hash from schema: %w", err)
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
