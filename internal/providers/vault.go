package providers

import (
	"errors"
	"time"

	"github.com/hashicorp/vault/api"
)

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
