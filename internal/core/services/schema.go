package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/json_schema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

type schema struct {
	repo   ports.SchemaRepository
	loader loader.DocumentLoader
}

// NewSchema is the schema service constructor
func NewSchema(repo ports.SchemaRepository, loader loader.DocumentLoader) *schema {
	return &schema{repo: repo, loader: loader}
}

// GetByID returns a domain.Schema by ID
func (s *schema) GetByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) (*domain.Schema, error) {
	schema, err := s.repo.GetByID(ctx, issuerDID, id)
	if errors.Is(err, repositories.ErrSchemaDoesNotExist) {
		return nil, ErrSchemaNotFound
	}
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// GetAll return all schemas in the database that matches the query string
func (s *schema) GetAll(ctx context.Context, issuerDID w3c.DID, query *string) ([]domain.Schema, error) {
	return s.repo.GetAll(ctx, issuerDID, query)
}

// ImportSchema process an schema url and imports into the system
func (s *schema) ImportSchema(ctx context.Context, did w3c.DID, req *ports.ImportSchemaRequest) (*domain.Schema, error) {
	remoteSchema, err := json_schema.Load(ctx, req.URL, s.loader)
	if err != nil {
		log.Error(ctx, "loading json_schema", "err", err, "json_schema", req.URL)
		return nil, ErrLoadingSchema
	}
	attributeNames, err := remoteSchema.Attributes()
	if err != nil {
		log.Error(ctx, "processing json_schema", "err", err, "json_schema", req.URL)
		return nil, ErrProcessSchema
	}

	hash, err := remoteSchema.SchemaHash(req.SType)
	if err != nil {
		log.Error(ctx, "hashing schema", "err", err, "json_schema", req.URL)
		return nil, ErrProcessSchema
	}

	schema := &domain.Schema{
		ID:          uuid.New(),
		IssuerDID:   did,
		URL:         req.URL,
		Type:        req.SType,
		Version:     req.Version,
		Title:       req.Title,
		Description: req.Description,
		Hash:        hash,
		Words:       attributeNames.SchemaAttrs(),
		CreatedAt:   time.Now(),
	}

	if err := s.repo.Save(ctx, schema); err != nil {
		log.Error(ctx, "saving imported schema", "err", err)
		return nil, err
	}
	return schema, nil
}
