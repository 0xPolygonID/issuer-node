package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const duplicatedEntryPGCode = "23505"

var (
	ErrSchemaDoesNotExist = errors.New("schema does not exist")   // ErrSchemaDoesNotExist schema does not exist
	ErrSchemaDuplicated   = errors.New("schema already imported") // ErrSchemaDuplicated schema duplicated
)

type dbSchema struct {
	ID          uuid.UUID
	IssuerID    string
	URL         string
	Type        string
	Version     string
	Title       *string
	Description *string
	Hash        string
	Words       string
	CreatedAt   time.Time
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
	const insertSchema = `INSERT INTO schemas (id, issuer_id, url, type,  hash,  words, created_at,version,title,description) VALUES($1, $2::text, $3::text, $4::text, $5::text, $6::text, $7, $8::text,$9::text,$10::text);`
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
		string(hash),
		r.toFullTextSearchDocument(s.Type, s.Words),
		s.CreatedAt,
		s.Version,
		s.Title,
		s.Description)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicatedEntryPGCode {
			return ErrSchemaDuplicated
		}

		return err
	}

	return nil
}

func (r *schema) toFullTextSearchDocument(sType string, attrs domain.SchemaWords) string {
	out := make([]string, 0, len(attrs)+1)
	out = append(out, sType)
	out = append(out, attrs...)
	return strings.Join(out, ", ")
}

// GetAll returns all the schemas that match any of the words that are included in the query string.
// For each word, it will search for attributes that start with it or include it following postgres full text search tokenization
func (r *schema) GetAll(ctx context.Context, issuerDID w3c.DID, query *string) ([]domain.Schema, error) {
	var err error
	var rows pgx.Rows
	sqlArgs := make([]interface{}, 0)
	sqlQuery := `SELECT id, issuer_id, url, type, words, hash, created_at,version,title,description
	FROM schemas
	WHERE issuer_id=$1`
	sqlArgs = append(sqlArgs, issuerDID.String())
	if query != nil && *query != "" {
		terms := tokenizeQuery(*query)
		sqlQuery += " AND (" + buildPartialQueryLikes("schemas.words", "OR", 1+len(sqlArgs), len(terms)) + ")"
		for _, term := range terms {
			sqlArgs = append(sqlArgs, term)
		}
	}
	sqlQuery += " ORDER BY created_at DESC"

	rows, err = r.conn.Pgx.Query(ctx, sqlQuery, sqlArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	schemaCol := make([]domain.Schema, 0)
	s := dbSchema{}
	for rows.Next() {
		if err := rows.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.Words, &s.Hash, &s.CreatedAt, &s.Version, &s.Title, &s.Description); err != nil {
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
func (r *schema) GetByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) (*domain.Schema, error) {
	const byID = `SELECT id, issuer_id, url, type, words, hash, created_at,version,title,description
		FROM schemas 
		WHERE issuer_id = $1 AND id=$2`

	s := dbSchema{}
	row := r.conn.Pgx.QueryRow(ctx, byID, issuerDID.String(), id)
	err := row.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.Words, &s.Hash, &s.CreatedAt, &s.Version, &s.Title, &s.Description)
	if err == pgx.ErrNoRows {
		return nil, ErrSchemaDoesNotExist
	}
	if err != nil {
		return nil, err
	}
	return toSchemaDomain(&s)
}

func toSchemaDomain(s *dbSchema) (*domain.Schema, error) {
	issuerDID, err := w3c.ParseDID(s.IssuerID)
	if err != nil {
		return nil, fmt.Errorf("parsing issuer DID from schema: %w", err)
	}
	schemaHash, err := core.NewSchemaHashFromHex(s.Hash)
	if err != nil {
		return nil, fmt.Errorf("parsing hash from schema: %w", err)
	}
	return &domain.Schema{
		ID:          s.ID,
		IssuerDID:   *issuerDID,
		URL:         s.URL,
		Type:        s.Type,
		Hash:        schemaHash,
		Words:       domain.SchemaWordsFromString(s.Words),
		CreatedAt:   s.CreatedAt,
		Version:     s.Version,
		Title:       s.Title,
		Description: s.Description,
	}, nil
}
