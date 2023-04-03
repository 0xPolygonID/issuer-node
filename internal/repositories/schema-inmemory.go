package repositories

import (
	"context"
	core "github.com/iden3/go-iden3-core"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

type schemaInMemory struct {
	schemas map[uuid.UUID]domain.Schema
}

// NewSchemaInMemory returns schemaRepository implemented in memory convenient for testing
func NewSchemaInMemory() *schemaInMemory {
	return &schemaInMemory{schemas: make(map[uuid.UUID]domain.Schema)}
}

func (s *schemaInMemory) Save(_ context.Context, schema *domain.Schema) error {
	s.schemas[schema.ID] = *schema
	return nil
}

func (s *schemaInMemory) GetByID(_ context.Context, _ core.DID, id uuid.UUID) (*domain.Schema, error) {
	if schema, found := s.schemas[id]; found {
		return &schema, nil
	}
	return nil, ErrSchemaDoesNotExist
}

// GetAll returns all. WARNING: query param will not work in the same way as DB repo
func (s *schemaInMemory) GetAll(_ context.Context, _ core.DID, _ *string) ([]domain.Schema, error) {
	schemas := make([]domain.Schema, len(s.schemas))
	i := 0
	for _, schema := range s.schemas {
		schemas[i] = schema
		i++
	}
	return schemas, nil
}
