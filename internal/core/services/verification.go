package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	"github.com/iden3/go-iden3-core/v2/w3c"
	protocol "github.com/iden3/iden3comm/v2/protocol"
	"github.com/jackc/pgtype"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// Scope is a property of VerificationQuery, and it's required for the authRequest
type Scope struct {
	CircuitId       string                  `json:"circuitId"`
	Id              uint32                  `json:"id"`
	Params          *map[string]interface{} `json:"params,omitempty"`
	Query           map[string]interface{}  `json:"query"`
	TransactionData *TransactionData        `json:"transactionData,omitempty"`
}

// TransactionData is a property of Scope, and it's required for the authRequest
type TransactionData struct {
	ChainID         int    `json:"chainID"`
	ContractAddress string `json:"contractAddress"`
	MethodID        string `json:"methodID"`
	Network         string `json:"network"`
}

const (
	stateTransitionDelay = time.Minute * 5
	defaultReason        = "for testing purposes"
	defaultBigIntBase    = 10
)

// VerificationService can verify responses to verification queries
type VerificationService struct {
	verifier *auth.Verifier
	repo     ports.VerificationRepository
}

// NewVerificationService creates a new instance of VerificationService
func NewVerificationService(verifier *auth.Verifier, repo ports.VerificationRepository) *VerificationService {
	return &VerificationService{verifier: verifier, repo: repo}
}

// CheckVerification checks if a verification response already exists for a given verification query ID and userDID.
// If no response exists, it returns the verification query.
func (vs *VerificationService) CheckVerification(ctx context.Context, issuerID w3c.DID, verificationQueryID uuid.UUID) (*domain.VerificationResponse, *domain.VerificationQuery, error) {
	query, err := vs.repo.Get(ctx, issuerID, verificationQueryID)
	if err != nil {
		if err == repositories.VerificationQueryNotFoundError {
			return nil, nil, fmt.Errorf("verification query not found: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to get verification query: %w", err)
	}

	response, err := vs.repo.GetVerificationResponse(ctx, verificationQueryID)
	if err == nil && response != nil {
		return response, nil, nil
	}

	return nil, query, nil
}

// SubmitVerificationResponse checks if a verification response passes a verify check and saves result
func (vs *VerificationService) SubmitVerificationResponse(ctx context.Context, verificationQueryID uuid.UUID, issuerID w3c.DID, token string, serverURL string) (*domain.VerificationResponse, error) {
	// check db for existing verification query
	var authRequest = protocol.AuthorizationRequestMessage{}
	verificationQuery, err := vs.repo.Get(ctx, issuerID, verificationQueryID)
	authRequest, err = getAuthRequestOffChain(verificationQuery, serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization request: %w", err)
	}

	// perform verification
	authRespMsg, err := vs.verifier.FullVerify(ctx, token,
		authRequest,
		pubsignals.WithAcceptedStateTransitionDelay(stateTransitionDelay))

	if err != nil {
		log.Error(ctx, "failed to verify", "verificationQueryID", verificationQueryID, "err", err)
		return nil, fmt.Errorf("failed to  verify token: %w", err)
	}

	// prepare to save result
	var jsonbResponse pgtype.JSONB
	err = jsonbResponse.Set(authRespMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to set JSONB value: %w", err)
	}

	response := domain.VerificationResponse{
		ID:                  uuid.New(), // Generate a new UUID for the response
		VerificationQueryID: verificationQueryID,
		UserDID:             authRespMsg.From,
		Response:            jsonbResponse,
		Pass:                true, // Assuming this field is present to indicate verification success
		CreatedAt:           time.Now(),
	}

	// Save the verification response to the repository
	responseID, err := vs.repo.AddResponse(ctx, verificationQueryID, response)
	if err != nil {
		return nil, fmt.Errorf("failed to add verification response: %w", err)
	}

	response.ID = responseID
	return &response, nil
}

func getAuthRequestOffChain(req *domain.VerificationQuery, serverURL string) (protocol.AuthorizationRequestMessage, error) {

	id := uuid.NewString()
	authReq := auth.CreateAuthorizationRequest(getReason(nil), req.IssuerDID, getUri(serverURL, req.IssuerDID, req.ID))
	authReq.ID = id
	authReq.ThreadID = id
	authReq.To = ""

	var scopes []Scope
	if req.Scope.Status == pgtype.Present {
		err := json.Unmarshal(req.Scope.Bytes, &scopes)
		if err != nil {
			return protocol.AuthorizationRequestMessage{}, err
		}
	}

	for _, scope := range scopes {
		mtpProofRequest := protocol.ZeroKnowledgeProofRequest{
			ID:        scope.Id,
			CircuitID: scope.CircuitId,
			Query:     scope.Query,
		}
		if scope.Params != nil {
			params, err := getParams(*scope.Params)
			if err != nil {
				return protocol.AuthorizationRequestMessage{}, err
			}

			mtpProofRequest.Params = params
		}
		authReq.Body.Scope = append(authReq.Body.Scope, mtpProofRequest)
	}
	return authReq, nil
}

func getReason(reason *string) string {
	if reason == nil {
		return "for testing purposes"
	}
	return *reason
}

func getParams(params map[string]interface{}) (map[string]interface{}, error) {
	val, ok := params["nullifierSessionID"]
	if !ok {
		return nil, errors.New("nullifierSessionID is empty")
	}

	nullifierSessionID := new(big.Int)
	if _, ok := nullifierSessionID.SetString(val.(string), defaultBigIntBase); !ok {
		return nil, errors.New("nullifierSessionID is not a valid big integer")
	}

	return map[string]interface{}{"nullifierSessionId": nullifierSessionID.String()}, nil
}

func getUri(serverURL string, issuerDID string, verificationQueryID uuid.UUID) string {
	path := fmt.Sprintf(`/v2/identities/%s/verification/callback?id=%s`, issuerDID, verificationQueryID)
	return fmt.Sprintf("%s%s", serverURL, path)
}
