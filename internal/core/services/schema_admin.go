package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
)

type schemaAdmin struct {
	repo          ports.SchemaRepository
	loaderFactory loader.Factory
}

// NewSchemaAdmin is the schemaAdmin service constructor
func NewSchemaAdmin(repo ports.SchemaRepository, lf loader.Factory) *schemaAdmin {
	return &schemaAdmin{repo: repo, loaderFactory: lf}
}

func (s *schemaAdmin) ImportSchema(ctx context.Context, did core.DID, url string, sType string) (*domain.Schema, error) {
	remoteSchema, err := jsonschema.Load(ctx, s.loaderFactory(url))
	if err != nil {
		log.Error(ctx, "loading jsonschema", "err", err, "jsonschema", url)
		return nil, ErrLoadingSchema
	}
	attributeNames, err := remoteSchema.AttributeNames()
	if err != nil {
		log.Error(ctx, "processing jsonschema", "err", err, "jsonschema", url)
		return nil, ErrProcessSchema
	}

	hash, err := remoteSchema.SchemaHash(sType)
	if err != nil {
		log.Error(ctx, "hashing schema", "err", err, "jsonschema", url)
		return nil, ErrProcessSchema
	}

	schema := &domain.Schema{
		ID:         uuid.New(),
		IssuerDID:  did,
		URL:        url,
		Type:       sType,
		Hash:       hash,
		Attributes: attributeNames.SchemaAttrs(),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Save(ctx, schema); err != nil {
		log.Error(ctx, "saving imported schema", "err", err)
		return nil, err
	}
	return schema, nil
}
