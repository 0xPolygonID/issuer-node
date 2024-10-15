package api

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/sqltools"
)

// PublishIdentityState - publish identity state on chain
func (s *Server) PublishIdentityState(ctx context.Context, request PublishIdentityStateRequestObject) (PublishIdentityStateResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return PublishIdentityState400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	publishedState, err := s.publisherGateway.PublishState(ctx, did)
	if err != nil {
		if errors.Is(err, gateways.ErrNoStatesToProcess) || errors.Is(err, gateways.ErrStateIsBeingProcessed) {
			return PublishIdentityState200JSONResponse{Message: err.Error()}, nil
		}

		var customErr *services.PublishingStateError
		if errors.As(err, &customErr) {
			return PublishIdentityState500JSONResponse{N500JSONResponse{Message: customErr.Error()}}, nil
		}

		return PublishIdentityState500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}

	return PublishIdentityState202JSONResponse{
		ClaimsTreeRoot:     publishedState.ClaimsTreeRoot,
		RevocationTreeRoot: publishedState.RevocationTreeRoot,
		RootOfRoots:        publishedState.RootOfRoots,
		State:              publishedState.State,
		TxID:               publishedState.TxID,
	}, nil
}

// RetryPublishState - retry to publish the current state if it failed previously.
func (s *Server) RetryPublishState(ctx context.Context, request RetryPublishStateRequestObject) (RetryPublishStateResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return RetryPublishState400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	publishedState, err := s.publisherGateway.RetryPublishState(ctx, did)
	if err != nil {
		log.Error(ctx, "error retrying the publishing the state", "err", err)
		if errors.Is(err, gateways.ErrStateIsBeingProcessed) || errors.Is(err, gateways.ErrNoFailedStatesToProcess) {
			return RetryPublishState400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
		}
		return RetryPublishState500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return RetryPublishState202JSONResponse{
		ClaimsTreeRoot:     publishedState.ClaimsTreeRoot,
		RevocationTreeRoot: publishedState.RevocationTreeRoot,
		RootOfRoots:        publishedState.RootOfRoots,
		State:              publishedState.State,
		TxID:               publishedState.TxID,
	}, nil
}

// GetStateTransactions - get state transactions
func (s *Server) GetStateTransactions(ctx context.Context, request GetStateTransactionsRequestObject) (GetStateTransactionsResponseObject, error) {
	filter, err := getStateTransitionsFilter(request)
	if err != nil {
		return GetStateTransactions400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetStateTransactions400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	states, total, err := s.identityService.GetStates(ctx, *did, filter)
	if err != nil {
		log.Error(ctx, "get state transactions", "err", err)
		return GetStateTransactions500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	return GetStateTransactions200JSONResponse(stateTransactionsPaginatedResponse(states, filter.Pagination, total)), nil
}

// GetStateStatus - get state status
func (s *Server) GetStateStatus(ctx context.Context, request GetStateStatusRequestObject) (GetStateStatusResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetStateStatus400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}
	pendingActions, err := s.identityService.HasUnprocessedAndFailedStatesByID(ctx, *did)
	if err != nil {
		log.Error(ctx, "get state status", "err", err)
		return GetStateStatus500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	return GetStateStatus200JSONResponse{PendingActions: pendingActions}, nil
}

func getStateTransitionsFilter(req GetStateTransactionsRequestObject) (request *ports.GetStateTransactionsRequest, err error) {
	const defaultFilter = "all"
	if req.Params.Page != nil && *req.Params.Page <= 0 {
		return nil, errors.New("page must be greater than 0")
	}
	filter := defaultFilter
	if req.Params.Filter != nil && !slices.Contains([]string{"all", "latest"}, strings.ToLower(string(*req.Params.Filter))) {
		return nil, errors.New("invalid filter")
	}
	orderBy := sqltools.OrderByFilters{}
	if req.Params.Sort != nil {
		for _, sortBy := range *req.Params.Sort {
			var err error
			field, desc := strings.CutPrefix(strings.TrimSpace(string(sortBy)), "-")
			switch GetStateTransactionsParamsSort(field) {
			case PublishDate:
				err = orderBy.Add(ports.StateTransitionsPublishDate, desc)
			case Status:
				err = orderBy.Add(ports.StateTransitionsStatus, desc)
			default:
				return nil, errors.New("wrong sort field")
			}
			if err != nil {
				return nil, errors.New("repeated sort by value field")
			}
		}
	}
	return ports.NewGetStateTransactionsRequest(filter, req.Params.Page, req.Params.MaxResults, orderBy), nil
}
