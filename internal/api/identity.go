package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// CreateIdentity is created identity controller
func (s *Server) CreateIdentity(ctx context.Context, request CreateIdentityRequestObject) (CreateIdentityResponseObject, error) {
	method := request.Body.DidMetadata.Method
	blockchain := request.Body.DidMetadata.Blockchain
	network := request.Body.DidMetadata.Network
	keyType := request.Body.DidMetadata.Type
	credentialStatusTypeRequest := request.Body.CredentialStatusType

	if keyType != "BJJ" && keyType != "ETH" {
		return CreateIdentity400JSONResponse{
			N400JSONResponse{
				Message: "Type must be BJJ or ETH",
			},
		}, nil
	}

	var credentialStatusType *verifiable.CredentialStatusType
	credentialStatusType, err := validateStatusType((*string)(credentialStatusTypeRequest))
	if err != nil {
		return CreateIdentity400JSONResponse{
			N400JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	rhsSettings, err := s.networkResolver.GetRhsSettingsForBlockchainAndNetwork(ctx, blockchain, network)
	if err != nil {
		return CreateIdentity400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("error getting reverse hash service settings: %s", err.Error())}}, nil
	}

	if !s.networkResolver.IsCredentialStatusTypeSupported(rhsSettings, *credentialStatusType) {
		log.Warn(ctx, "unsupported credential status type", "req", request)
		return CreateIdentity400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("Credential Status Type '%s' is not supported by the issuer", *credentialStatusType)}}, nil
	}

	identity, err := s.identityService.Create(ctx, s.cfg.ServerUrl, &ports.DIDCreationOptions{
		Method:               core.DIDMethod(method),
		Network:              core.NetworkID(network),
		Blockchain:           core.Blockchain(blockchain),
		KeyType:              kms.KeyType(keyType),
		AuthCredentialStatus: *credentialStatusType,
		DisplayName:          request.Body.DisplayName,
	})
	if err != nil {
		if errors.Is(err, services.ErrWrongDIDMetada) {
			return CreateIdentity400JSONResponse{
				N400JSONResponse{
					Message: err.Error(),
				},
			}, nil
		}

		if errors.Is(err, kms.ErrPermissionDenied) {
			var message string
			if s.cfg.KeyStore.VaultUserPassAuthEnabled {
				message = "Issuer Node cannot connect with Vault. Please check the value of ISSUER_VAULT_USERPASS_AUTH_PASSWORD variable."
			} else {
				message = `Issuer Node cannot connect with Vault. Please check the value of ISSUER_KEY_STORE_TOKEN variable.`
			}

			log.Info(ctx, message+". More information in this link: https://docs.privado.id/docs/issuer/vault-auth/")
			return CreateIdentity403JSONResponse{
				N403JSONResponse{
					Message: message,
				},
			}, nil
		}

		var customErr *services.PublishingStateError
		if errors.As(err, &customErr) {
			return CreateIdentity400JSONResponse{
				N400JSONResponse{
					Message: customErr.Error(),
				},
			}, nil
		}

		return nil, err
	}

	var responseAddress *string
	if identity.Address != nil && *identity.Address != "" {
		responseAddress = identity.Address
	}

	return CreateIdentity201JSONResponse{
		Identifier:  &identity.Identifier,
		DisplayName: identity.DisplayName,
		State: &IdentityState{
			BlockNumber:        identity.State.BlockNumber,
			BlockTimestamp:     identity.State.BlockTimestamp,
			ClaimsTreeRoot:     identity.State.ClaimsTreeRoot,
			CreatedAt:          TimeUTC(identity.State.CreatedAt),
			ModifiedAt:         TimeUTC(identity.State.ModifiedAt),
			PreviousState:      identity.State.PreviousState,
			RevocationTreeRoot: identity.State.RevocationTreeRoot,
			RootOfRoots:        identity.State.RootOfRoots,
			State:              identity.State.State,
			Status:             string(identity.State.Status),
			TxID:               identity.State.TxID,
		},
		Address:              responseAddress,
		KeyType:              identity.KeyType,
		Balance:              nil,
		CredentialStatusType: CreateIdentityResponseCredentialStatusType(identity.AuthCoreClaimRevocationStatus.Type),
	}, nil
}

// validateStatusType - validate credential status type.
// If credentialStatusTypeRequest is nil or empty, it will return Iden3commRevocationStatusV1
func validateStatusType(credentialStatusTypeRequest *string) (*verifiable.CredentialStatusType, error) {
	var credentialStatusType verifiable.CredentialStatusType
	if credentialStatusTypeRequest != nil && *credentialStatusTypeRequest != "" {
		allowedCredentialStatuses := []string{string(verifiable.Iden3commRevocationStatusV1), string(verifiable.Iden3ReverseSparseMerkleTreeProof), string(verifiable.Iden3OnchainSparseMerkleTreeProof2023)}
		if !slices.Contains(allowedCredentialStatuses, *credentialStatusTypeRequest) {
			return nil, errors.New(fmt.Sprintf("Invalid Credential Status Type '%s'. Allowed Iden3commRevocationStatusV1.0, Iden3ReverseSparseMerkleTreeProof or Iden3OnchainSparseMerkleTreeProof2023.", *credentialStatusTypeRequest))
		}
		credentialStatusType = (verifiable.CredentialStatusType)(*credentialStatusTypeRequest)
	} else {
		credentialStatusType = verifiable.Iden3commRevocationStatusV1
	}
	return &credentialStatusType, nil
}

// UpdateIdentity is update identity display name controller
func (s *Server) UpdateIdentity(ctx context.Context, request UpdateIdentityRequestObject) (UpdateIdentityResponseObject, error) {
	userDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "update identity.. Parsing did", "err", err)
		return UpdateIdentity400JSONResponse{
			N400JSONResponse{
				Message: "invalid did",
			},
		}, err
	}

	err = s.identityService.UpdateIdentityDisplayName(ctx, *userDID, request.Body.DisplayName)
	if err != nil {
		log.Error(ctx, "update identity. updating display name", "err", err)
		return UpdateIdentity400JSONResponse{
			N400JSONResponse{
				Message: "invalid identity",
			},
		}, err
	}

	return UpdateIdentity200JSONResponse{Message: "Identity display name updated"}, nil
}

// GetIdentities is the controller to get identities
func (s *Server) GetIdentities(ctx context.Context, request GetIdentitiesRequestObject) (GetIdentitiesResponseObject, error) {
	var err error
	var response GetIdentities200JSONResponse
	identities, err := s.identityService.Get(ctx)
	if err != nil {
		return GetIdentities500JSONResponse{N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	partsLength := 4
	for _, identity := range identities {
		did, err := w3c.ParseDID(identity.Identifier)
		if err != nil {
			return GetIdentities500JSONResponse{N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}

		var authBjjCredStatus *GetIdentitiesResponseCredentialStatusType
		authClaim, _ := s.claimService.GetAuthClaim(ctx, did)
		if authClaim != nil {
			credentialStatus, _ := authClaim.GetCredentialStatus()

			if credentialStatus != nil {
				credStatusType := GetIdentitiesResponseCredentialStatusType(credentialStatus.Type)
				authBjjCredStatus = &credStatusType
			}
		}

		items := strings.Split(identity.Identifier, ":")
		if len(items) < partsLength {
			return GetIdentities500JSONResponse{N500JSONResponse{
				Message: "Invalid identity",
			}}, nil
		}

		response = append(response, GetIdentitiesResponse{
			Identifier:           identity.Identifier,
			Method:               items[1],
			Blockchain:           items[2],
			Network:              items[3],
			CredentialStatusType: authBjjCredStatus,
			DisplayName:          identity.DisplayName,
		})
	}

	return response, nil
}

// GetIdentityDetails is the controller to get identity details
func (s *Server) GetIdentityDetails(ctx context.Context, request GetIdentityDetailsRequestObject) (GetIdentityDetailsResponseObject, error) {
	userDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "get identity details. Parsing did", "err", err)
		return GetIdentityDetails400JSONResponse{
			N400JSONResponse{
				Message: "invalid did",
			},
		}, err
	}

	identity, err := s.identityService.GetByDID(ctx, *userDID)
	if err != nil {
		log.Error(ctx, "get identity details. Getting identity", "err", err)
		if errors.Is(err, repositories.ErrIdentityNotFound) {
			return GetIdentityDetails400JSONResponse{
				N400JSONResponse{
					Message: "identity not found",
				},
			}, nil
		}

		return GetIdentityDetails500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, err
	}

	var balance *big.Int
	if identity.KeyType == string(kms.KeyTypeEthereum) {
		did, err := w3c.ParseDID(identity.Identifier)
		if err != nil {
			log.Error(ctx, "get identity details. Parsing did", "err", err)
			return GetIdentityDetails400JSONResponse{N400JSONResponse{Message: "invalid did"}}, nil
		}
		balance, err = s.accountService.GetBalanceByDID(ctx, did)
		if err != nil {
			log.Error(ctx, "get identity details. Getting balance", "err", err)
			return GetIdentityDetails500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
		}
		identity.Balance = balance
	}

	var responseBalance *string
	if balance != nil {
		responseBalance = common.ToPointer(balance.String())
	}

	var responseAddress *string
	if identity.Address != nil && *identity.Address != "" {
		responseAddress = identity.Address
	}

	response := GetIdentityDetails200JSONResponse{
		Identifier: &identity.Identifier,
		State: &IdentityState{
			BlockNumber:        identity.State.BlockNumber,
			BlockTimestamp:     identity.State.BlockTimestamp,
			ClaimsTreeRoot:     identity.State.ClaimsTreeRoot,
			CreatedAt:          TimeUTC(identity.State.CreatedAt),
			ModifiedAt:         TimeUTC(identity.State.ModifiedAt),
			PreviousState:      identity.State.PreviousState,
			RevocationTreeRoot: identity.State.RevocationTreeRoot,
			RootOfRoots:        identity.State.RootOfRoots,
			State:              identity.State.State,
			Status:             string(identity.State.Status),
			TxID:               identity.State.TxID,
		},
		KeyType:              identity.KeyType,
		DisplayName:          identity.DisplayName,
		Address:              responseAddress,
		Balance:              responseBalance,
		CredentialStatusType: GetIdentityDetailsResponseCredentialStatusType(identity.AuthCoreClaimRevocationStatus.Type),
	}

	return response, nil
}
