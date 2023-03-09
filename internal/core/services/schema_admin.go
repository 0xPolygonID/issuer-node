package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

type schemaAdmin struct {
	repo ports.SchemaRepository
}

// NewSchemaAdmin is the schemaAdmin service constructor
func NewSchemaAdmin(repo ports.SchemaRepository) *schemaAdmin {
	return &schemaAdmin{repo: repo}
}

func (s *schemaAdmin) ImportSchema(ctx context.Context, did core.DID, url string, sType string) (*domain.Schema, error) {
	// TODO:
	hash := core.SchemaHash{}
	attrs := domain.SchemaAttrs{}

	schema := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  did,
		URL:        url,
		Type:       sType,
		Hash:       hash,
		Attributes: attrs,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Save(ctx, schema); err != nil {
		return nil, err
	}
	return schema, nil
}
