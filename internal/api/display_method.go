package api

import (
	"context"
	"errors"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/sqltools"
)

// CreateDisplayMethod - Create a new display method handler
func (s *Server) CreateDisplayMethod(ctx context.Context, request CreateDisplayMethodRequestObject) (CreateDisplayMethodResponseObject, error) {
	identityDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "Invalid identifier", "err", err)
		return CreateDisplayMethod400JSONResponse{N400JSONResponse{Message: "Invalid identifier"}}, nil
	}

	if request.Body.Name == "" {
		log.Error(ctx, "Name is required")
		return CreateDisplayMethod400JSONResponse{N400JSONResponse{Message: "name is required"}}, nil
	}
	if request.Body.Url == "" {
		log.Error(ctx, "Url is required")
		return CreateDisplayMethod400JSONResponse{N400JSONResponse{Message: "url is required"}}, nil
	}

	id, err := s.displayMethodService.Save(ctx, *identityDID, request.Body.Name, request.Body.Url, request.Body.Type)
	if err != nil {
		log.Error(ctx, "Error saving display method", "err", err)
		if errors.Is(err, repositories.DisplayMethodDuplicateNameError) {
			return CreateDisplayMethod400JSONResponse{N400JSONResponse{Message: "Name already exist"}}, nil
		}
		return CreateDisplayMethod500JSONResponse{N500JSONResponse{Message: "Error saving display method"}}, nil
	}
	return CreateDisplayMethod201JSONResponse{
		Id: id.String(),
	}, nil
}

// GetDisplayMethod - Get display method handler
func (s *Server) GetDisplayMethod(ctx context.Context, request GetDisplayMethodRequestObject) (GetDisplayMethodResponseObject, error) {
	identityDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "Invalid identifier", "err", err)
		return GetDisplayMethod400JSONResponse{N400JSONResponse{Message: "Invalid identifier"}}, nil
	}
	displayMethod, err := s.displayMethodService.GetByID(ctx, *identityDID, request.Id)
	if err != nil {
		log.Error(ctx, "Error getting display method", "err", err)
		if errors.Is(err, repositories.DisplayMethodNotFoundErr) {
			return GetDisplayMethod404JSONResponse{N404JSONResponse{Message: "Display method not found"}}, nil
		}
		return GetDisplayMethod500JSONResponse{N500JSONResponse{Message: "Error getting display method"}}, nil
	}

	return GetDisplayMethod200JSONResponse{
		Id:   displayMethod.ID,
		Name: displayMethod.Name,
		Url:  displayMethod.URL,
		Type: displayMethod.Type,
	}, nil
}

// GetAllDisplayMethods - Get all display methods handler
func (s *Server) GetAllDisplayMethods(ctx context.Context, request GetAllDisplayMethodsRequestObject) (GetAllDisplayMethodsResponseObject, error) {
	identityDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "Invalid identifier", "err", err)
		return GetAllDisplayMethods400JSONResponse{N400JSONResponse{Message: "Invalid identifier"}}, nil
	}

	filter, err := getDisplayMethodFilter(request)
	if err != nil {
		log.Error(ctx, "Error getting filter", "err", err)
		return GetAllDisplayMethods400JSONResponse{N400JSONResponse{Message: "Error getting filter"}}, nil
	}

	displayMethods, total, err := s.displayMethodService.GetAll(ctx, *identityDID, *filter)
	if err != nil {
		log.Error(ctx, "Error getting display methods", "err", err)
		return GetAllDisplayMethods500JSONResponse{N500JSONResponse{Message: "Error getting display methods"}}, nil
	}

	var response GetAllDisplayMethods200JSONResponse
	items := make([]DisplayMethodEntity, 0, len(displayMethods))
	for _, displayMethod := range displayMethods {
		items = append(items, getDisplayMethodResponseObject(&displayMethod))
	}

	response.Items = items
	response.Meta = PaginatedMetadata{
		Total:      total,
		Page:       filter.Page,
		MaxResults: filter.MaxResults,
	}

	return response, nil
}

// UpdateDisplayMethod - update display method handler
func (s *Server) UpdateDisplayMethod(ctx context.Context, request UpdateDisplayMethodRequestObject) (UpdateDisplayMethodResponseObject, error) {
	identityDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "Invalid identifier", "err", err)
		return UpdateDisplayMethod400JSONResponse{N400JSONResponse{Message: "Invalid identifier"}}, nil
	}

	if request.Body.Name != nil && *request.Body.Name == "" {
		log.Error(ctx, "Name cannot be empty")
		return UpdateDisplayMethod400JSONResponse{N400JSONResponse{Message: "name cannot be empty"}}, nil
	}
	if request.Body.Url != nil && *request.Body.Url == "" {
		log.Error(ctx, "Url is required")
		return UpdateDisplayMethod400JSONResponse{N400JSONResponse{Message: "url cannot be empty"}}, nil
	}

	_, err = s.displayMethodService.Update(ctx, *identityDID, request.Id, request.Body.Name, request.Body.Url, request.Body.Type)
	if err != nil {
		log.Error(ctx, "Error updating display method", "err", err)
		if errors.Is(err, repositories.DisplayMethodNotFoundErr) {
			return UpdateDisplayMethod404JSONResponse{N404JSONResponse{Message: "Display method not found"}}, nil
		}
		if errors.Is(err, repositories.DisplayMethodDuplicateNameError) {
			return UpdateDisplayMethod400JSONResponse{N400JSONResponse{Message: "Duplicated name display method"}}, nil
		}
		return UpdateDisplayMethod500JSONResponse{N500JSONResponse{Message: "Error updating display method"}}, nil
	}

	return UpdateDisplayMethod200JSONResponse{
		Message: "Display method updated",
	}, nil
}

// DeleteDisplayMethod - delete display method handler
func (s *Server) DeleteDisplayMethod(ctx context.Context, request DeleteDisplayMethodRequestObject) (DeleteDisplayMethodResponseObject, error) {
	identityDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "Invalid identifier", "err", err)
		return DeleteDisplayMethod400JSONResponse{N400JSONResponse{Message: "Invalid identifier"}}, nil
	}
	err = s.displayMethodService.Delete(ctx, *identityDID, request.Id)
	if err != nil {
		log.Error(ctx, "Error deleting display method", "err", err)
		if errors.Is(err, repositories.DisplayMethodNotFoundErr) {
			return DeleteDisplayMethod404JSONResponse{N404JSONResponse{Message: "Display method not found"}}, nil
		}
		if errors.Is(err, services.DefaultDisplayMethodErr) {
			log.Error(ctx, "Cannot delete default display method", "err", err)
			return DeleteDisplayMethod400JSONResponse{N400JSONResponse{Message: "Cannot delete default display method"}}, nil
		}
		return DeleteDisplayMethod500JSONResponse{N500JSONResponse{Message: "Error deleting display method"}}, nil
	}

	return DeleteDisplayMethod200JSONResponse{
		Message: "Display method deleted",
	}, nil
}

// getDisplayMethodResponseObject - creates a response object from a display method
func getDisplayMethodResponseObject(displayMethod *domain.DisplayMethod) DisplayMethodEntity {
	return DisplayMethodEntity{
		Id:   displayMethod.ID,
		Name: displayMethod.Name,
		Url:  displayMethod.URL,
		Type: displayMethod.Type,
	}
}

// getDisplayMethodFilter - creates a display method filter from the request
func getDisplayMethodFilter(req GetAllDisplayMethodsRequestObject) (*ports.DisplayMethodFilter, error) {
	filter := ports.DisplayMethodFilter{}
	filter.MaxResults = 50
	if req.Params.MaxResults != nil {
		if *req.Params.MaxResults <= 0 {
			filter.MaxResults = 10
		} else {
			filter.MaxResults = *req.Params.MaxResults
		}
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
			switch GetAllDisplayMethodsParamsSort(field) {
			case CreatedAt:
				err = orderBy.Add(ports.DisplayMethodCreatedAtFilterField, desc)
			case Name:
				err = orderBy.Add(ports.DisplayMethodNameFilterField, desc)
			case Type:
				err = orderBy.Add(ports.DisplayMethodTypeFilterField, desc)
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
