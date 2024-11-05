package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// ImportSchema is the UI endpoint to import schema metadata
func (s *Server) ImportSchema(ctx context.Context, request ImportSchemaRequestObject) (ImportSchemaResponseObject, error) {
	req := request.Body
	if err := guardImportSchemaReq(req); err != nil {
		log.Debug(ctx, "Importing schema bad request", "err", err, "req", req)
		return ImportSchema400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("bad request: %s", err.Error())}}, nil
	}
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return ImportSchema400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	iReq := ports.NewImportSchemaRequest(req.Url, req.SchemaType, req.Title, req.Version, req.Description)
	schema, err := s.schemaService.ImportSchema(ctx, *issuerDID, iReq)
	if err != nil {
		log.Error(ctx, "Importing schema", "err", err, "req", req)
		return ImportSchema500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return ImportSchema201JSONResponse{Id: schema.ID.String()}, nil
}

// GetSchema is the UI endpoint that searches and schema by Id and returns it.
func (s *Server) GetSchema(ctx context.Context, request GetSchemaRequestObject) (GetSchemaResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetSchema400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	schema, err := s.schemaService.GetByID(ctx, *issuerDID, request.Id)
	if errors.Is(err, services.ErrSchemaNotFound) {
		log.Error(ctx, "schema not found", "id", request.Id)
		return GetSchema404JSONResponse{N404JSONResponse{Message: "schema not found"}}, nil
	}
	if err != nil {
		log.Error(ctx, "loading schema", "err", err, "id", request.Id)
	}
	return GetSchema200JSONResponse(schemaResponse(schema)), nil
}

// GetSchemas returns the list of schemas that match the request.Params.Query filter. If param query is nil it will return all
func (s *Server) GetSchemas(ctx context.Context, request GetSchemasRequestObject) (GetSchemasResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetSchemas400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	filter := ports.SchemasFilter{
		Query: request.Params.Query,
	}

	filter.MaxResults = 50
	if request.Params.MaxResults != nil {
		if *request.Params.MaxResults <= 0 {
			filter.MaxResults = 10
		} else {
			filter.MaxResults = *request.Params.MaxResults
		}
	}
	if request.Params.Page == nil {
		filter.Page = uint(1)
	} else {
		filter.Page = *request.Params.Page
	}

	col, total, err := s.schemaService.GetAll(ctx, *issuerDID, filter)
	if err != nil {
		log.Error(ctx, "loading schemas", "err", err)
		return GetSchemas500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	response := GetSchemas200JSONResponse{
		Items: schemaCollectionResponse(col),
		Meta: PaginatedMetadata{
			Total:      total,
			Page:       filter.Page,
			MaxResults: filter.MaxResults,
		},
	}

	return response, nil
}

func guardImportSchemaReq(req *ImportSchemaJSONRequestBody) error {
	if req == nil {
		return errors.New("empty body")
	}
	if strings.TrimSpace(req.Url) == "" {
		return errors.New("empty url")
	}
	if strings.TrimSpace(req.SchemaType) == "" {
		return errors.New("empty type")
	}
	if _, err := url.ParseRequestURI(req.Url); err != nil {
		return fmt.Errorf("parsing url: %w", err)
	}
	return nil
}
