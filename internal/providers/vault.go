package providers

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/polygonid/sh-id-platform/internal/log"
)

const (
	didMountPath = "kv"
	secretPath   = "did"
)

// DidNotFound error
var DidNotFound = errors.New("did not found in vault")

// HTTPClientTimeout http client timeout TODO: move to config
const HTTPClientTimeout = 10 * time.Second

// NewVaultClient checks vault configuration and creates new vault client
func NewVaultClient(address, token string) (*api.Client, error) {
	if address == "" {
		return nil, errors.New("vault address is not specified")
	}
	if token == "" {
		return nil, errors.New("vault access token is not specified")
	}

	config := api.DefaultConfig()
	config.Address = address
	config.HttpClient.Timeout = HTTPClientTimeout

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return client, nil
}

// GetDID gets did from vault
func GetDID(ctx context.Context, vaultCli *api.Client) (string, error) {
	did, err := vaultCli.KVv2(didMountPath).Get(ctx, secretPath)
	if err != nil {
		log.Error(ctx, "error getting did from vault", "error", err)
		return "", DidNotFound
	}

	if did.Data["did"] == nil {
		log.Info(ctx, "did not found in vault")
		return "", DidNotFound
	}

	didToReturn, ok := did.Data["did"].(string)
	if !ok {
		log.Error(ctx, "error casting did to string")
		return "", DidNotFound
	}
	return didToReturn, nil
}

// SaveDID saves did to vault
func SaveDID(ctx context.Context, vaultCli *api.Client, did string) error {
	_, err := vaultCli.KVv2(didMountPath).Put(ctx, secretPath, map[string]interface{}{
		"did": did,
	})
	if err != nil {
		log.Error(ctx, "error saving did to vault", "error", err)
		return err
	}

	log.Info(ctx, "did saved to vault")
	return nil
}
