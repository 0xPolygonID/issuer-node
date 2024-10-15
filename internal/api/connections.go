package api

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	jsonSuite "github.com/iden3/go-schema-processor/v2/json"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/sql_tools"
)

// GetConnections returns the list of connections of a determined issuer
func (s *Server) GetConnections(ctx context.Context, request GetConnectionsRequestObject) (GetConnectionsResponseObject, error) {
	filter, err := getConnectionsFilter(request)
	if err != nil {
		return GetConnections400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetConnections400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	conns, total, err := s.connectionsService.GetAllByIssuerID(ctx, *issuerDID, filter)
	if err != nil {
		log.Error(ctx, "get connection request", "err", err)
		return GetConnections500JSONResponse{N500JSONResponse{"Unexpected error while retrieving connections"}}, nil
	}

	resp, err := connectionsPaginatedResponse(conns, filter.Pagination, total)
	if err != nil {
		log.Error(ctx, "get connection request invalid claim format", "err", err)
		return GetConnections500JSONResponse{N500JSONResponse{"Unexpected error while retrieving connections"}}, nil

	}
	return GetConnections200JSONResponse(resp), nil
}

// CreateConnection creates a connection
func (s *Server) CreateConnection(ctx context.Context, request CreateConnectionRequestObject) (CreateConnectionResponseObject, error) {
	// validate did document
	issuerDoc, err := json.Marshal(request.Body.IssuerDoc)
	if err != nil {
		log.Debug(ctx, "invalid issuer did document object", "err", err, "req", request)
		return CreateConnection400JSONResponse{N400JSONResponse{"invalid issuer did document"}}, nil
	}

	userDoc, err := json.Marshal(request.Body.UserDoc)
	if err != nil {
		log.Debug(ctx, "invalid user did document object", "err", err, "req", request)
		return CreateConnection400JSONResponse{N400JSONResponse{"invalid user did document"}}, nil
	}
	validator := jsonSuite.Validator{}
	if checkJSONIsNotNull(issuerDoc) {
		err := validator.ValidateData(issuerDoc, []byte(verifiable.DIDDocumentJSONSchema))
		if err != nil {
			log.Debug(ctx, "invalid issuer did document", "err", err, "req", request)
			return CreateConnection400JSONResponse{N400JSONResponse{"invalid issuer did document"}}, nil
		}
	}

	if checkJSONIsNotNull(userDoc) {
		err := validator.ValidateData(userDoc, []byte(verifiable.DIDDocumentJSONSchema))
		if err != nil {
			log.Debug(ctx, "invalid user did document", "err", err, "req", request)
			return CreateConnection400JSONResponse{N400JSONResponse{"invalid user did document"}}, nil
		}
	}

	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Debug(ctx, "invalid issuer did", "err", err, "req", request)
		return CreateConnection400JSONResponse{N400JSONResponse{"invalid issuer did"}}, nil
	}

	userDID, err := w3c.ParseDID(request.Body.UserDID)
	if err != nil {
		log.Debug(ctx, "invalid user did", "err", err, "req", request)
		return CreateConnection400JSONResponse{N400JSONResponse{"invalid user did"}}, nil
	}

	conn := &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  *issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  issuerDoc,
		UserDoc:    userDoc,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	err = s.connectionsService.Create(ctx, conn)
	if err != nil {
		log.Debug(ctx, "create connection error", "err", err, "req", request)
		return CreateConnection500JSONResponse{N500JSONResponse{"There was an error creating the connection"}}, nil
	}

	return CreateConnection201JSONResponse{}, nil
}

// DeleteConnection deletes a connection
func (s *Server) DeleteConnection(ctx context.Context, request DeleteConnectionRequestObject) (DeleteConnectionResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return DeleteConnection400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	req := ports.NewDeleteRequest(request.Id, request.Params.DeleteCredentials, request.Params.RevokeCredentials)
	if req.RevokeCredentials {
		err := s.claimService.RevokeAllFromConnection(ctx, req.ConnID, *issuerDID)
		if err != nil {
			log.Error(ctx, "delete connection, revoking credentials", "err", err, "req", request.Id.String())
			return DeleteConnection500JSONResponse{N500JSONResponse{"There was an error revoking the credentials of the given connection"}}, nil
		}
	}

	err = s.connectionsService.Delete(ctx, request.Id, req.DeleteCredentials, *issuerDID)
	if err != nil {
		if errors.Is(err, services.ErrConnectionDoesNotExist) {
			log.Info(ctx, "delete connection, non existing conn", "err", err, "req", request.Id.String())
			return DeleteConnection400JSONResponse{N400JSONResponse{"The given connection does not exist"}}, nil
		}
		log.Error(ctx, "delete connection", "err", err, "req", request.Id.String())
		return DeleteConnection500JSONResponse{N500JSONResponse{deleteConnection500Response(req.DeleteCredentials, req.RevokeCredentials)}}, nil
	}

	return DeleteConnection200JSONResponse{Message: deleteConnectionResponse(req.DeleteCredentials, req.RevokeCredentials)}, nil
}

// GetConnection returns a connection with its related credentials
func (s *Server) GetConnection(ctx context.Context, request GetConnectionRequestObject) (GetConnectionResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetConnection400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	conn, err := s.connectionsService.GetByIDAndIssuerID(ctx, request.Id, *issuerDID)
	if err != nil {
		if errors.Is(err, services.ErrConnectionDoesNotExist) {
			return GetConnection400JSONResponse{N400JSONResponse{"The given connection does not exist"}}, nil
		}
		log.Debug(ctx, "get connection internal server error", "err", err, "req", request)
		return GetConnection500JSONResponse{N500JSONResponse{"There was an error retrieving the connection"}}, nil
	}

	filter := &ports.ClaimsFilter{
		Subject: conn.UserDID.String(),
	}
	credentials, _, err := s.claimService.GetAll(ctx, *issuerDID, filter)
	if err != nil && !errors.Is(err, services.ErrCredentialNotFound) {
		log.Debug(ctx, "get connection internal server error retrieving credentials", "err", err, "req", request)
		return GetConnection500JSONResponse{N500JSONResponse{"There was an error retrieving the connection"}}, nil
	}

	resp, err := connectionResponse(conn, credentials)
	if err != nil {
		log.Error(ctx, "get connection internal server error converting credentials to w3c", "err", err)
		return GetConnection500JSONResponse{N500JSONResponse{"There was an error parsing the credential of the given connection"}}, nil
	}

	return GetConnection200JSONResponse(resp), nil
}

// DeleteConnectionCredentials deletes all the credentials of the given connection
func (s *Server) DeleteConnectionCredentials(ctx context.Context, request DeleteConnectionCredentialsRequestObject) (DeleteConnectionCredentialsResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return DeleteConnectionCredentials400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if err := s.connectionsService.DeleteCredentials(ctx, request.Id, *issuerDID); err != nil {
		log.Error(ctx, "delete connection request", err, "req", request)
		return DeleteConnectionCredentials500JSONResponse{N500JSONResponse{"There was an error deleting the credentials of the given connection"}}, nil
	}

	return DeleteConnectionCredentials200JSONResponse{Message: "Credentials of the connection successfully deleted"}, nil
}

// RevokeConnectionCredentials revoke all the non revoked credentials of the given connection
func (s *Server) RevokeConnectionCredentials(ctx context.Context, request RevokeConnectionCredentialsRequestObject) (RevokeConnectionCredentialsResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return RevokeConnectionCredentials400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if err := s.claimService.RevokeAllFromConnection(ctx, request.Id, *issuerDID); err != nil {
		log.Error(ctx, "revoke connection credentials", "err", err, "req", request)
		return RevokeConnectionCredentials500JSONResponse{N500JSONResponse{"There was an error revoking the credentials of the given connection"}}, nil
	}

	return RevokeConnectionCredentials202JSONResponse{Message: "Credentials revocation request sent"}, nil
}

func getConnectionsFilter(req GetConnectionsRequestObject) (*ports.NewGetAllConnectionsRequest, error) {
	if req.Params.Page != nil && *req.Params.Page <= 0 {
		return nil, errors.New("page must be greater than 0")
	}
	orderBy := sql_tools.OrderByFilters{}
	if req.Params.Sort != nil {
		for _, sortBy := range *req.Params.Sort {
			var err error
			field, desc := strings.CutPrefix(strings.TrimSpace(string(sortBy)), "-")
			switch GetConnectionsParamsSort(field) {
			case GetConnectionsParamsSortCreatedAt:
				err = orderBy.Add(ports.ConnectionsCreatedAt, desc)
			case GetConnectionsParamsSortUserID:
				err = orderBy.Add(ports.ConnectionsUserID, desc)
			default:
				return nil, errors.New("wrong sort by value")
			}
			if err != nil {
				return nil, errors.New("repeated sort by value field")
			}
		}
	}
	return ports.NewGetAllRequest(req.Params.Credentials, req.Params.Query, req.Params.Page, req.Params.MaxResults, orderBy), nil
}

func checkJSONIsNotNull(message []byte) bool {
	return message != nil && string(message) != "null"
}
