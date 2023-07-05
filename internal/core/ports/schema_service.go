package ports

import (
	"context"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaService defines the methods that Schema manager will expose.
type SchemaService interface {
	ImportSchema(ctx context.Context, issuerDID core.DID, req *ImportSchemaRequest) (*domain.Schema, error)
	GetByID(ctx context.Context, issuerDID core.DID, id uuid.UUID) (*domain.Schema, error)
	GetAll(ctx context.Context, issuerDID core.DID, query *string) ([]domain.Schema, error)
}

// ImportSchemaRequest defines the request for importing a schema
type ImportSchemaRequest struct {
	URL         string
	SType       string
	Title       string
	Description string
	Version     string
}

// NewImportSchemaRequest creates a new ImportSchemaRequest
func NewImportSchemaRequest(url string, stype string, title string, version string, description string) *ImportSchemaRequest {
	return &ImportSchemaRequest{
		URL:         url,
		SType:       stype,
		Title:       title,
		Description: description,
		Version:     version,
	}
}
