package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	const insertSchema = `INSERT INTO schemas (id, issuer_id, url, type, attributes, hash, ts_words, created_at) VALUES($1, $2::text, $3::text, $4::text, $5::text, $6::text, to_tsvector($7::text), $8);`
	hash, err := s.Hash.MarshalText()
	if err != nil {
		return err
	}
	_, err = r.conn.Pgx.Exec(
		ctx,
		insertSchema,
		s.ID,
		s.IssuerDID.String(),
		s.URL,
		s.Type,
		s.Attributes.String(),
		string(hash),
		r.toFullTextSearchDocument(s.Type, s.Attributes),
		s.CreatedAt)
	return err
}

func (r *schema) toFullTextSearchDocument(sType string, attrs domain.SchemaAttrs) string {
	var sb strings.Builder
	sb.WriteString(sType + " ")
	sb.WriteString(" ")
	for _, attr := range attrs {
		sb.WriteString(attr + " ")
	}
	return sb.String()
}

// GetAll returns all the schemas that match any of the words that are included in the query string.
// For each word, it will search for attributes that start with it or include it following postgres full text search tokenization
func (r *schema) GetAll(ctx context.Context, query *string) ([]domain.Schema, error) {
	const all = `SELECT id, issuer_id, url, type, attributes, hash, created_at
	FROM schemas
	ORDER BY created_at DESC`
	const allFTS = `
SELECT id, issuer_id, url, type, attributes, hash, created_at 
FROM schemas 
WHERE ts_words @@ to_tsquery($1)
ORDER BY created_at DESC`
	var err error
	var rows pgx.Rows

	if query != nil && *query != "" {
		rows, err = r.conn.Pgx.Query(ctx, allFTS, fullTextSearchQuery(*query, " | "))
	} else {
		rows, err = r.conn.Pgx.Query(ctx, all)
	}
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
	return schemaCol, rows.Err()
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
