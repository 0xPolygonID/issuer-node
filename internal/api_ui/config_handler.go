package api_ui

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
			Key:   "ISSUER_API_UI_SERVER_URL",
			Value: s.cfg.APIUI.ServerURL,
		},

		KeyValue{
			Key:   "ISSUER_API_BLOCKCHAIN",
			Value: s.cfg.APIUI.IdentityBlockchain,
		},

		KeyValue{
			Key:   "ISSUER_API_METHOD",
			Value: s.cfg.APIUI.IdentityMethod,
		},

		KeyValue{
			Key:   "ISSUER_API_NETWORK",
			Value: s.cfg.APIUI.IdentityNetwork,
		},

		KeyValue{
			Key:   "ISSUER_API_UI_ISSUER_DID",
			Value: s.cfg.APIUI.IssuerDID.String(),
		},

		KeyValue{
			Key:   "ISSUER_API_IPFS_GATEWAY_URL",
			Value: s.cfg.IPFS.GatewayURL,
		},
	}

	return variables, nil
}
