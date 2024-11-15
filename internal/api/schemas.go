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
	"github.com/polygonid/sh-id-platform/internal/sqltools"
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

	filter, err := getFilter(request)
	if err != nil {
		log.Error(ctx, "getting filter", "err", err)
		return GetSchemas400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}

	col, total, err := s.schemaService.GetAll(ctx, *issuerDID, *filter)
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

func getFilter(req GetSchemasRequestObject) (*ports.SchemasFilter, error) {
	filter := ports.SchemasFilter{}
	filter.MaxResults = 50
	if req.Params.MaxResults != nil {
		if *req.Params.MaxResults <= 0 {
			filter.MaxResults = 10
		} else {
			filter.MaxResults = *req.Params.MaxResults
		}
	}

	if req.Params.Query != nil {
		filter.Query = req.Params.Query
	}

	filter.Page = uint(1)
	if req.Params.Page != nil {
		filter.Page = *req.Params.Page
	}

	orderBy := sqltools.OrderByFilters{}
	if req.Params.Sort != nil {
		for _, sortBy := range *req.Params.Sort {
			var err error
			field, desc := strings.CutPrefix(strings.TrimSpace(string(sortBy)), "-")
			switch GetSchemasParamsSort(field) {
			case ImportDate:
				err = orderBy.Add(ports.SchemaImportDateFilterField, desc)
			case SchemaType:
				err = orderBy.Add(ports.SchemaTypeFilterField, desc)
			case SchemaVersion:
				err = orderBy.Add(ports.SchemaVersionFilterField, desc)
			default:
				return nil, errors.New("wrong sort field")
			}
			if err != nil {
				return nil, errors.New("repeated sort by value field")
			}
		}
	}
	filter.OrderBy = orderBy
	return &filter, nil
}
