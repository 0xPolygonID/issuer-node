package api

import (
	"context"
	"errors"
	"fmt"

	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/gateways"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// CreateIdentity is created identity controller
func (s *Server) CreateIdentity(ctx context.Context, request CreateIdentityRequestObject) (CreateIdentityResponseObject, error) {
	method := request.Body.DidMetadata.Method
	blockchain := request.Body.DidMetadata.Blockchain
	network := request.Body.DidMetadata.Network
	keyType := request.Body.DidMetadata.Type
	authBJJCredentialStatus := request.Body.DidMetadata.AuthBJJCredentialStatus

	if keyType != "BJJ" && keyType != "ETH" {
		return CreateIdentity400JSONResponse{
			N400JSONResponse{
				Message: "Type must be BJJ or ETH",
			},
		}, nil
	}

	rhsSettings, err := s.networkResolver.GetRhsSettingsForBlockchainAndNetwork(ctx, blockchain, network)
	if err != nil {
		return CreateIdentity400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("error getting reverse hash service settings: %s", err.Error())}}, nil
	}

	if authBJJCredentialStatus == nil {
		authBJJCredentialStatus = (*string)(&rhsSettings.DefaultAuthBJJCredentialStatus)
	}

	identity, err := s.identityService.Create(ctx, s.cfg.ServerUrl, &ports.DIDCreationOptions{
		Method:                  core.DIDMethod(method),
		Network:                 core.NetworkID(network),
		Blockchain:              core.Blockchain(blockchain),
		KeyType:                 kms.KeyType(keyType),
		AuthBJJCredentialStatus: (verifiable.CredentialStatusType)(*authBJJCredentialStatus),
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
			if s.cfg.VaultUserPassAuthEnabled {
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

		return nil, err
	}

	return CreateIdentity201JSONResponse{
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
		Address: identity.Address,
	}, nil
}

// GetIdentities is the controller to get identities
func (s *Server) GetIdentities(ctx context.Context, request GetIdentitiesRequestObject) (GetIdentitiesResponseObject, error) {
	var response GetIdentities200JSONResponse
	var err error
	response, err = s.identityService.Get(ctx)
	if err != nil {
		return GetIdentities500JSONResponse{N500JSONResponse{
			Message: err.Error(),
		}}, nil
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
		return GetIdentityDetails500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, err
	}

	if identity.KeyType == string(kms.KeyTypeEthereum) {
		did, err := w3c.ParseDID(identity.Identifier)
		if err != nil {
			log.Error(ctx, "get identity details. Parsing did", "err", err)
			return GetIdentityDetails400JSONResponse{N400JSONResponse{Message: "invalid did"}}, nil
		}
		balance, err := s.accountService.GetBalanceByDID(ctx, did)
		if err != nil {
			log.Error(ctx, "get identity details. Getting balance", "err", err)
			return GetIdentityDetails500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
		}
		identity.Balance = balance
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
	}

	if identity.Address != nil && *identity.Address != "" {
		response.Address = identity.Address
	}

	if identity.Balance != nil {
		response.Balance = common.ToPointer(identity.Balance.String())
	}

	return response, nil
}

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
