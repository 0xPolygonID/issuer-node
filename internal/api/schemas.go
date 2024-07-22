package api

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

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
	iReq := ports.NewImportSchemaRequest(req.Url, req.SchemaType, req.Title, req.Version, req.Description)
	schema, err := s.schemaService.ImportSchema(ctx, s.cfg.APIUI.IssuerDID, iReq)
	if err != nil {
		log.Error(ctx, "Importing schema", "err", err, "req", req)
		return ImportSchema500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return ImportSchema201JSONResponse{Id: schema.ID.String()}, nil
}

// GetSchema is the UI endpoint that searches and schema by Id and returns it.
func (s *Server) GetSchema(ctx context.Context, request GetSchemaRequestObject) (GetSchemaResponseObject, error) {
	schema, err := s.schemaService.GetByID(ctx, s.cfg.APIUI.IssuerDID, request.Id)
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
	col, err := s.schemaService.GetAll(ctx, s.cfg.APIUI.IssuerDID, request.Params.Query)
	if err != nil {
		log.Error(ctx, "loading schemas", "err", err)
		return GetSchemas500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return GetSchemas200JSONResponse(schemaCollectionResponse(col)), nil
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
