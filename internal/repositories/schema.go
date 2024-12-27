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

var (
	ErrSchemaDoesNotExist    = errors.New("schema does not exist")    // ErrSchemaDoesNotExist schema does not exist
	ErrDuplicated            = errors.New("schema already imported")  // ErrDuplicated schema duplicated
	ErrDisplayMethodNotFound = errors.New("display method not found") // ErrDisplayMethodNotFound display method not found
)

type dbSchema struct {
	ID              uuid.UUID
	IssuerID        string
	URL             string
	Type            string
	ContextURL      string
	Version         string
	Title           *string
	Description     *string
	Hash            string
	Words           string
	DisplayMethodID *uuid.UUID
	CreatedAt       time.Time
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
	const insertSchema = `INSERT INTO schemas (id, issuer_id, url, type,  context_url, hash,  words, created_at, version, title, description, display_method_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8,$9,$10,$11,$12);`
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
		s.ContextURL,
		string(hash),
		r.toFullTextSearchDocument(s.Type, s.Words),
		s.CreatedAt,
		s.Version,
		s.Title,
		s.Description,
		s.DisplayMethodID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateViolationErrorCode {
			return ErrDuplicated
		}

		if pgErr.Code == foreignKeyViolationErrorCode {
			return ErrDisplayMethodNotFound
		}

		return err
	}
	return nil
}

func (r *schema) Update(ctx context.Context, schema *domain.Schema) error {
	const updateSchema = `
UPDATE schemas 
SET issuer_id=$2, url=$3, type=$4, context_url=$5, hash=$6,  words=$7, created_at=$8, version=$9, title=$10, description=$11, display_method_id=$12
WHERE schemas.id = $1;`
	hash, err := schema.Hash.MarshalText()
	if err != nil {
		return err
	}
	res, err := r.conn.Pgx.Exec(
		ctx,
		updateSchema,
		schema.ID,
		schema.IssuerDID.String(),
		schema.URL,
		schema.Type,
		schema.ContextURL,
		string(hash),
		r.toFullTextSearchDocument(schema.Type, schema.Words),
		schema.CreatedAt,
		schema.Version,
		schema.Title,
		schema.Description,
		schema.DisplayMethodID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationErrorCode {
			return DisplayMethodAssignedErr
		}
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrSchemaDoesNotExist
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
	sqlQuery := `SELECT id, issuer_id, url, type, context_url, words, hash, created_at,version,title,description,display_method_id
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
		if err := rows.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.ContextURL, &s.Words, &s.Hash, &s.CreatedAt, &s.Version, &s.Title, &s.Description, &s.DisplayMethodID); err != nil {
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
	const byID = `SELECT id, issuer_id, url, type, context_url, words, hash, created_at,version,title,description,display_method_id
		FROM schemas 
		WHERE issuer_id = $1 AND id=$2`

	s := dbSchema{}
	row := r.conn.Pgx.QueryRow(ctx, byID, issuerDID.String(), id)
	err := row.Scan(&s.ID, &s.IssuerID, &s.URL, &s.Type, &s.ContextURL, &s.Words, &s.Hash, &s.CreatedAt, &s.Version, &s.Title, &s.Description, &s.DisplayMethodID)
	if errors.Is(err, pgx.ErrNoRows) {
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
		ID:              s.ID,
		IssuerDID:       *issuerDID,
		URL:             s.URL,
		Type:            s.Type,
		ContextURL:      s.ContextURL,
		Hash:            schemaHash,
		Words:           domain.SchemaWordsFromString(s.Words),
		CreatedAt:       s.CreatedAt,
		Version:         s.Version,
		Title:           s.Title,
		Description:     s.Description,
		DisplayMethodID: s.DisplayMethodID,
	}, nil
}
