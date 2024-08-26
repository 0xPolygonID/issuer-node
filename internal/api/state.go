package api

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/log"
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
	const (
		defaultPage       = uint(1)
		defaultMaxResults = uint(10)
		defaultFilter     = "all"
	)
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetStateTransactions400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	page := defaultPage
	if request.Params.Page != nil {
		page = *request.Params.Page
	}

	maxResults := defaultMaxResults
	if request.Params.MaxResults != nil {
		maxResults = *request.Params.MaxResults
	}

	filter := defaultFilter
	if request.Params.Filter != nil {
		filter = string(*request.Params.Filter)
	}

	states, err := s.identityService.GetStates(ctx, *did, page, maxResults, filter)
	if err != nil {
		log.Error(ctx, "get state transactions", "err", err)
		return GetStateTransactions500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	return GetStateTransactions200JSONResponse(stateTransactionsResponse(states)), nil
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
