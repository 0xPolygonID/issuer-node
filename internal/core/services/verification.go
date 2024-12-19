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

	"github.com/polygonid/sh-id-platform/internal/cache"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// Scope is a property of VerificationQuery, and it's required for the authRequest
type scope struct {
	CircuitId       string           `json:"circuitId"`
	Id              uint32           `json:"id"`
	Params          json.RawMessage  `json:"params,omitempty"`
	Query           json.RawMessage  `json:"query"`
	TransactionData *transactionData `json:"transactionData,omitempty"`
}

// TransactionData is a property of Scope, and it's required for the authRequest
type transactionData struct {
	ChainID         int    `json:"chainID"`
	ContractAddress string `json:"contractAddress"`
	MethodID        string `json:"methodID"`
	Network         string `json:"network"`
}

const (
	stateTransitionDelay               = time.Minute * 5
	defaultReason                      = "for testing purposes"
	defaultBigIntBase                  = 10
	defaultExpiration    time.Duration = 0
)

var errVerificationKeyNotFound = errors.New("authRequest not found in the cache")

// VerificationService can verify responses to verification queries
type VerificationService struct {
	networkResolver *network.Resolver
	store           cache.Cache
	repo            ports.VerificationRepository
	verifier        *auth.Verifier
}

// NewVerificationService creates a new instance of VerificationService
func NewVerificationService(networkResolver *network.Resolver, store cache.Cache, repo ports.VerificationRepository, verifier *auth.Verifier) *VerificationService {
	return &VerificationService{
		networkResolver: networkResolver,
		store:           store,
		verifier:        verifier,
		repo:            repo,
	}
}

// CreateVerificationQuery  creates and saves a new verification query in the database.
func (vs *VerificationService) CreateVerificationQuery(ctx context.Context, issuerID w3c.DID, chainID int, skipCheckRevocation bool, scopes []map[string]interface{}, serverURL string) (*domain.VerificationQuery, error) {
	var scopeJSON pgtype.JSONB
	err := scopeJSON.Set(scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to set scope JSON: %w", err)
	}

	verificationQuery := domain.VerificationQuery{
		ID:                  uuid.New(),
		IssuerDID:           issuerID.String(),
		ChainID:             chainID,
		SkipCheckRevocation: skipCheckRevocation,
		Scope:               scopeJSON,
		CreatedAt:           time.Now(),
	}

	queryID, err := vs.repo.Save(ctx, issuerID, verificationQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to save verification query: %w", err)
	}

	verificationQuery.ID = queryID

	authRequest, err := vs.getAuthRequestOffChain(&verificationQuery, serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth request: %w", err)
	}

	if err := vs.store.Set(ctx, vs.key(queryID), authRequest, defaultExpiration); err != nil {
		log.Error(ctx, "error storing verification query request", "id", queryID.String(), "error", err)
		return nil, err
	}

	return &verificationQuery, nil
}

// GetVerificationStatus checks if a verification response already exists for a given verification query ID and userDID.
// If no response exists, it returns the verification query.
func (vs *VerificationService) GetVerificationStatus(ctx context.Context, issuerID w3c.DID, verificationQueryID uuid.UUID) (*domain.VerificationResponse, *domain.VerificationQuery, error) {
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
	// check cache for existing authRequest
	var authRequest protocol.AuthorizationRequestMessage
	if found := vs.store.Get(ctx, vs.key(verificationQueryID), &authRequest); !found {
		log.Error(ctx, "authRequest not found in the cache", "id", verificationQueryID.String())
		return nil, errVerificationKeyNotFound
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

func (s *VerificationService) key(id uuid.UUID) string {
	return "issuer-node:qr-code:" + id.String()
}

func (vs *VerificationService) getAuthRequestOffChain(req *domain.VerificationQuery, serverURL string) (protocol.AuthorizationRequestMessage, error) {
	id := uuid.NewString()
	authReq := auth.CreateAuthorizationRequest(vs.getReason(nil), req.IssuerDID, vs.getUri(serverURL, req.IssuerDID, req.ID))
	authReq.ID = id
	authReq.ThreadID = id
	authReq.To = ""

	var scopes []scope
	if req.Scope.Status == pgtype.Present {
		err := json.Unmarshal(req.Scope.Bytes, &scopes)
		if err != nil {
			return protocol.AuthorizationRequestMessage{}, err
		}
	}

	for _, scope := range scopes {
		var query map[string]interface{}
		err := json.Unmarshal(scope.Query, &query)
		if err != nil {
			return protocol.AuthorizationRequestMessage{}, err
		}

		mtpProofRequest := protocol.ZeroKnowledgeProofRequest{
			ID:        scope.Id,
			CircuitID: scope.CircuitId,
			Query:     query,
		}
		if scope.Params != nil {
			var paramsObj map[string]interface{}
			err := json.Unmarshal(scope.Params, &paramsObj)
			if err != nil {
				return protocol.AuthorizationRequestMessage{}, err
			}
			params, err := vs.getParams(paramsObj)
			if err != nil {
				return protocol.AuthorizationRequestMessage{}, err
			}

			mtpProofRequest.Params = params
		}
		authReq.Body.Scope = append(authReq.Body.Scope, mtpProofRequest)
	}
	return authReq, nil
}

func (vs *VerificationService) getReason(reason *string) string {
	if reason == nil {
		return "for testing purposes"
	}
	return *reason
}

func (vs *VerificationService) getParams(params map[string]interface{}) (map[string]interface{}, error) {
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

func (vs *VerificationService) getUri(serverURL string, issuerDID string, verificationQueryID uuid.UUID) string {
	path := fmt.Sprintf(`/v2/identities/%s/verification/callback?id=%s`, issuerDID, verificationQueryID)
	return fmt.Sprintf("%s%s", serverURL, path)
}
