package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// CreateVerification a VerificationQuery using verificationService
func (s *Server) CreateVerification(ctx context.Context, request CreateVerificationRequestObject) (CreateVerificationResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreateVerification400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	verificationQuery, err := s.verificationService.Create(ctx, *issuerDID, request.Body.ChainId, request.Body.SkipRevocationCheck, request.Body.Scopes, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "creating verification query", "err", err)
		return CreateVerification500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	return CreateVerification200JSONResponse{
		Message: verificationQuery.ID.String(),
	}, nil
}

// CheckVerification returns CheckVerificationResponse or Provided Query
func (s *Server) CheckVerification(ctx context.Context, request CheckVerificationRequestObject) (CheckVerificationResponseObject, error) {
	// validation
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CheckVerification400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	verificationQueryID, err := uuid.Parse(request.Params.Id.String())
	if err != nil {
		log.Error(ctx, "invalid verification query ID", "err", err)
		return CheckVerification400JSONResponse{N400JSONResponse{Message: "invalid verification query ID"}}, nil
	}

	// Use the VerificationService to check if there's an existing response or get the query
	response, query, err := s.verificationService.Check(ctx, *issuerDID, verificationQueryID)
	if err != nil {
		log.Error(ctx, "checking verification", "err", err)
		return CheckVerification500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	if response != nil {
		return CheckVerification200JSONResponse{union: json.RawMessage(fmt.Sprintf(`{"status": "validated", "response": %v}`, response))}, nil
	}

	if query != nil {
		return CheckVerification200JSONResponse{union: json.RawMessage(fmt.Sprintf(`{"query": %v}`, query))}, nil
	}

	return CheckVerification404JSONResponse{N404JSONResponse{Message: "Verification query not found"}}, nil
}

// SubmitVerificationResponse returns a VerificationResponse
func (s *Server) SubmitVerificationResponse(ctx context.Context, request SubmitVerificationResponseRequestObject) (SubmitVerificationResponseResponseObject, error) {
	// Unmarshal user DID and response data
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "submitting verification query response.. Parsing did", "err", err)
		return SubmitVerificationResponse400JSONResponse{
			N400JSONResponse{
				Message: "invalid did",
			},
		}, err
	}
	token := request.Body

	// Submit the verification response using VerificationService
	response, err := s.verificationService.Submit(ctx, request.Params.Id, *issuerDID, *token, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "failed to submit verification response", "err", err)
		return SubmitVerificationResponse500JSONResponse{
			N500JSONResponse{
				Message: "failed to submit verification response",
			},
		}, nil
	}

	// Construct and return the response status
	return SubmitVerificationResponse200JSONResponse{
		Status: "submitted",
		Pass:   response.Pass,
	}, nil
}
