package api

import (
	"context"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// GetConfig - GetEthClient configuration
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
			Key:   "ISSUER_KEY_STORE_TOKEN",
			Value: token,
		},

		KeyValue{
			Key:   "ISSUER_REDIS_URL",
			Value: s.cfg.Cache.Url,
		},

		KeyValue{
			Key:   "ISSUER_API_IPFS_GATEWAY_URL",
			Value: s.cfg.IPFS.GatewayURL,
		},
	}

	return variables, nil
}
