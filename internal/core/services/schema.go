package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

type schema struct {
	repo                 ports.SchemaRepository
	loader               loader.DocumentLoader
	displayMethodService ports.DisplayMethodService
}

// NewSchema is the schema service constructor
func NewSchema(repo ports.SchemaRepository, loader loader.DocumentLoader, displayMethodService ports.DisplayMethodService) *schema {
	return &schema{repo: repo, loader: loader, displayMethodService: displayMethodService}
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
	schemaID := schema.ID
	schema, err = s.fixSchemaContext(ctx, schema)
	if err != nil {
		log.Error(ctx, "fixing schema context", "err", err, "schema", schemaID)
		return nil, fmt.Errorf("fixing schema context: %w", err)
	}
	return schema, nil
}

// fixSchemaContext updates the schema context url if it is empty. This will happen in old installations
// that did not have the context url stored in the database
// There is no action in DB if the context url is already stored
func (s *schema) fixSchemaContext(ctx context.Context, schema *domain.Schema) (*domain.Schema, error) {
	if schema.ContextURL == "" {
		remoteSchema, err := jsonschema.Load(ctx, schema.URL, s.loader)
		if err != nil {
			log.Error(ctx, "loading jsonschema", "err", err, "jsonschema", schema.URL)
			return nil, ErrLoadingSchema
		}
		contextUrl, err := remoteSchema.JSONLdContext()
		if err != nil {
			log.Error(ctx, "getting jsonld context", "err", err, "jsonschema", schema.URL)
			return nil, ErrProcessSchema
		}
		schema.ContextURL = contextUrl
		if err := s.repo.Update(ctx, schema); err != nil {
			return nil, fmt.Errorf("updating schema: %w", err)
		}
	}
	return schema, nil
}

// GetAll return all schemas in the database that matches the query string
func (s *schema) GetAll(ctx context.Context, issuerDID w3c.DID, query *string) ([]domain.Schema, error) {
	return s.repo.GetAll(ctx, issuerDID, query)
}

// ImportSchema process an schema url and imports into the system
func (s *schema) ImportSchema(ctx context.Context, did w3c.DID, req *ports.ImportSchemaRequest) (*domain.Schema, error) {
	remoteSchema, err := jsonschema.Load(ctx, req.URL, s.loader)
	if err != nil {
		log.Error(ctx, "loading jsonschema", "err", err, "jsonschema", req.URL)
		return nil, ErrLoadingSchema
	}
	attributeNames, err := remoteSchema.Attributes()
	if err != nil {
		log.Error(ctx, "processing jsonschema", "err", err, "jsonschema", req.URL)
		return nil, ErrProcessSchema
	}

	hash, err := remoteSchema.SchemaHash(req.SType)
	if err != nil {
		log.Error(ctx, "hashing schema", "err", err, "jsonschema", req.URL)
		return nil, ErrProcessSchema
	}
	contextUrl, err := remoteSchema.JSONLdContext()
	if err != nil {
		log.Error(ctx, "getting jsonld context", "err", err, "jsonschema", req.URL)
		return nil, ErrProcessSchema
	}

	if req.DisplayMethodID != nil {
		_, err := s.displayMethodService.GetByID(ctx, did, *req.DisplayMethodID)
		if err != nil {
			log.Error(ctx, "getting display method", "err", err)
			return nil, ErrDisplayMethodNotFound
		}
	}

	schema := &domain.Schema{
		ID:              uuid.New(),
		IssuerDID:       did,
		URL:             req.URL,
		Type:            req.SType,
		ContextURL:      contextUrl,
		Version:         req.Version,
		Title:           req.Title,
		Description:     req.Description,
		Hash:            hash,
		Words:           attributeNames.SchemaAttrs(),
		DisplayMethodID: req.DisplayMethodID,
		CreatedAt:       time.Now(),
	}

	if err := s.repo.Save(ctx, schema); err != nil {
		log.Error(ctx, "saving imported schema", "err", err)
		return nil, err
	}
	return schema, nil
}

// Update updates a schema
func (s *schema) Update(ctx context.Context, schema *domain.Schema) error {
	if schema.DisplayMethodID != nil {
		_, err := s.displayMethodService.GetByID(ctx, schema.IssuerDID, *schema.DisplayMethodID)
		if err != nil {
			log.Error(ctx, "getting display method", "err", err)
			return ErrDisplayMethodNotFound
		}
	}
	schemaInDatabase, err := s.repo.GetByID(ctx, schema.IssuerDID, schema.ID)
	if err != nil {
		return err
	}
	schemaInDatabase.DisplayMethodID = schema.DisplayMethodID
	return s.repo.Save(ctx, schemaInDatabase)
}
