package providers

import (
	"context"
	"errors"
	"time"

	vault "github.com/hashicorp/vault/api"
	auth2 "github.com/hashicorp/vault/api/auth/userpass"

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

// Config vault configuration
// If UserPassAuthEnabled is true, then vault client will be created with userpass auth and Pass must be provided
type Config struct {
	Address             string
	UserPassAuthEnabled bool
	Token               string
	Pass                string
}

// VaultClient checks vault configuration and creates new vault client
func VaultClient(ctx context.Context, cfg Config) (*vault.Client, error) {
	var vaultCli *vault.Client
	var err error
	if cfg.UserPassAuthEnabled {
		log.Info(ctx, "Vault userpass auth enabled")
		if cfg.Pass == "" {
			log.Error(ctx, "Vault userpass auth enabled but password not provided")
			return nil, errors.New("Vault userpass auth enabled but password not provided")
		}
		vaultCli, err = newVaultClientWithUserPassAuth(ctx, cfg.Address, cfg.Pass)
		if err != nil {
			log.Error(ctx, "cannot init vault client with userpass auth: ", "err", err)
			return nil, err
		}
	} else {
		log.Info(ctx, "Vault userpass auth not enabled")
		if cfg.Token == "" {
			log.Error(ctx, "Vault userpass auth not enabled but token not provided")
			return nil, errors.New("Vault userpass auth not enabled but token not provided")
		}
		vaultCli, err = newVaultClientWithToken(cfg.Address, cfg.Token)
		if err != nil {
			log.Error(ctx, "cannot init vault client: ", "err", err)
			return nil, err
		}
	}

	return vaultCli, nil
}

// newVaultClientWithToken checks vault configuration and creates new vault client
func newVaultClientWithToken(address, token string) (*vault.Client, error) {
	if address == "" {
		return nil, errors.New("vault address is not specified")
	}
	if token == "" {
		return nil, errors.New("vault access token is not specified")
	}

	config := vault.DefaultConfig()
	config.Address = address
	config.HttpClient.Timeout = HTTPClientTimeout

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(token)
	return client, nil
}

// newVaultClientWithUserPassAuth checks vault configuration and creates new vault client with userpass auth
func newVaultClientWithUserPassAuth(ctx context.Context, address string, pass string) (*vault.Client, error) {
	config := vault.DefaultConfig()
	config.Address = address
	config.HttpClient.Timeout = HTTPClientTimeout

	client, err := vault.NewClient(config)
	if err != nil {
		log.Error(ctx, "error creating vault client with userpass auth", "error", err)
		return nil, err
	}

	user := "issuernode"

	userPass, err := auth2.NewUserpassAuth(user, &auth2.Password{
		FromString: pass,
	})
	if err != nil {
		log.Error(ctx, "error creating userpass auth", "error", err)
		return nil, err
	}

	secret, err := client.Auth().Login(ctx, userPass)
	if err != nil {
		log.Error(ctx, "error logging in to vault with userpass auth", "error", err)
		return nil, err
	}

	log.Info(ctx, "successfully logged in to vault with userpass auth", "token", secret.Auth.ClientToken)
	return client, nil
}

// GetDID gets did from vault
func GetDID(ctx context.Context, vaultCli *vault.Client) (string, error) {
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
func SaveDID(ctx context.Context, vaultCli *vault.Client, did string) error {
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
