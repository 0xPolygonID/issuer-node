package api

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// GetConfig - Get configuration
func (s *Server) GetConfig(_ context.Context, _ GetConfigRequestObject) (GetConfigResponseObject, error) {
	token := common.ReplaceCharacters(s.cfg.KeyStore.Token)
	variables := GetConfig200JSONResponse{
		KeyValue{
			Key:   "ISSUER_SERVER_URL",
			Value: s.cfg.ServerUrl,
		},
		KeyValue{
			Key:   "ISSUER_DATABASE_URL",
			Value: s.cfg.Database.URL,
		},
		KeyValue{
			Key:   "ISSUER_ETHEREUM_URL",
			Value: s.cfg.Ethereum.URL,
		},

		KeyValue{
			Key:   "ISSUER_ETHEREUM_CONTRACT_ADDRESS",
			Value: s.cfg.Ethereum.ContractAddress,
		},

		KeyValue{
			Key:   "ISSUER_ETHEREUM_RESOLVER_PREFIX",
			Value: s.cfg.Ethereum.ResolverPrefix,
		},

		KeyValue{
			Key:   "ISSUER_KEY_STORE_TOKEN",
			Value: token,
		},

		KeyValue{
			Key:   "ISSUER_REDIS_URL",
			Value: s.cfg.Cache.RedisUrl,
		},

		KeyValue{
			Key:   "ISSUER_API_IPFS_GATEWAY_URL",
			Value: s.cfg.IPFS.GatewayURL,
		},

		KeyValue{
			Key:   "ISSUER_CREDENTIAL_STATUS_RHS_MODE",
			Value: string(s.cfg.CredentialStatus.RHSMode),
		},
	}

	if s.cfg.CredentialStatus.RHSMode == "OnChain" {
		variables = append(variables, KeyValue{
			Key:   "ISSUER_CREDENTIAL_STATUS_ONCHAIN_TREE_STORE_SUPPORTED_CONTRACT",
			Value: s.cfg.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract,
		})

		variables = append(variables, KeyValue{
			Key:   "ISSUER_CREDENTIAL_STATUS_RHS_CHAIN_ID",
			Value: s.cfg.CredentialStatus.OnchainTreeStore.ChainID,
		})
	}

	if s.cfg.CredentialStatus.RHSMode == "OffChain" {
		variables = append(variables, KeyValue{
			Key:   "ISSUER_CREDENTIAL_STATUS_RHS_URL",
			Value: s.cfg.CredentialStatus.RHS.URL,
		})
	}

	return variables, nil
}
