package repositories

import (
	"context"

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

func (s *schemaInMemory) GetById(_ context.Context, id uuid.UUID) (*domain.Schema, error) {
	if schema, found := s.schemas[id]; found {
		return &schema, nil
	}
	return nil, ErrSchemaDoesNotExist
}
