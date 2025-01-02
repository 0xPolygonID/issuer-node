package api

import (
	"context"
	"encoding/json"
	"fmt"

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

	verificationQuery, err := s.verificationService.CreateVerificationQuery(ctx, *issuerDID, request.Body.ChainId, request.Body.SkipRevocationCheck, request.Body.Scopes, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "creating verification query", "err", err)
		return CreateVerification500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}

	return CreateVerification201JSONResponse{
		VerificationQueryId: verificationQuery.ID.String(),
	}, nil
}

// CheckVerification returns CheckVerificationResponse or Provided Query
func (s *Server) CheckVerification(ctx context.Context, request CheckVerificationRequestObject) (CheckVerificationResponseObject, error) {
	// Parse and validate issuer DID
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CheckVerification400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	verificationQueryID := request.Params.Id

	// Use the VerificationService to check for existing response or query
	response, query, err := s.verificationService.GetVerificationStatus(ctx, *issuerDID, verificationQueryID)
	if err != nil {
		// if error is not found, return 404
		if err.Error() == "verification query not found" {
			log.Error(ctx, "checking verification, not found", "err", err, "issuerDID", issuerDID, "id", verificationQueryID)
			return CheckVerification404JSONResponse{N404JSONResponse{Message: "Verification query not found"}}, nil
		} else {
			log.Error(ctx, "checking verification", "err", err, "issuerDID", issuerDID, "id", verificationQueryID)
			return CheckVerification500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
		}
	}

	// Check if a response exists
	if response != nil {
		var jsonResponse map[string]interface{}

		// Marshal the interface{} value to JSON bytes
		responseBytes, err := json.Marshal(response.Response.Get())
		if err != nil {
			return CheckVerification500JSONResponse{
				N500JSONResponse{Message: fmt.Sprintf("failed to marshal response: %v", err)},
			}, nil
		}

		// Unmarshal the JSON bytes into map[string]interface{}
		if err := json.Unmarshal(responseBytes, &jsonResponse); err != nil {
			return CheckVerification500JSONResponse{
				N500JSONResponse{Message: fmt.Sprintf("failed to unmarshal response: %v", err)},
			}, nil
		}

		return CheckVerification200JSONResponse{
			VerificationResponse: &VerificationResponse{
				VerificationScopeId: response.VerificationQueryID.String(),
				UserDid:             response.UserDID,
				Response:            jsonResponse, // Safely populated JSON object
				Pass:                response.Pass,
			},
		}, nil
	}

	// Check if a query exists
	if query != nil {
		var scopeJSON map[string]interface{}

		// Marshal the interface{} value to JSON bytes
		scopeBytes, err := json.Marshal(query.Scope.Get())
		if err != nil {
			return CheckVerification500JSONResponse{
				N500JSONResponse{Message: fmt.Sprintf("failed to marshal scope: %v", err)},
			}, nil
		}

		// Unmarshal the JSON bytes into map[string]interface{}
		if err := json.Unmarshal(scopeBytes, &scopeJSON); err != nil {
			return CheckVerification500JSONResponse{
				N500JSONResponse{Message: fmt.Sprintf("failed to unmarshal scope: %v", err)},
			}, nil
		}

		return CheckVerification200JSONResponse{
			VerificationQueryRequest: &VerificationQueryRequest{
				VerificationQueryId: query.ID.String(),
				Scopes:              scopeJSON, // Safely populated JSON object
			},
		}, nil
	}

	// Return 404 if neither response nor query exists
	return CheckVerification404JSONResponse{N404JSONResponse{Message: "Verification query not found"}}, nil
}

// SubmitVerificationResponse returns a VerificationResponse
func (s *Server) SubmitVerificationResponse(ctx context.Context, request SubmitVerificationResponseRequestObject) (SubmitVerificationResponseResponseObject, error) {
	// Unmarshal user DID and response data
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "submitting verification query response.. Parsing did", "err", err, "issuerDID", issuerDID)
		return SubmitVerificationResponse400JSONResponse{
			N400JSONResponse{
				Message: "invalid did",
			},
		}, err
	}
	token := request.Body

	// Submit the verification response using VerificationService
	response, err := s.verificationService.SubmitVerificationResponse(ctx, request.Params.Id, *issuerDID, *token, s.cfg.ServerUrl)
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
		Status: Submitted,
		Pass:   response.Pass,
	}, nil
}
