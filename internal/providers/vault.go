package providers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
	auth2 "github.com/hashicorp/vault/api/auth/userpass"

	"github.com/polygonid/sh-id-platform/internal/log"
)

const (
	didMountPath = "kv"
	secretPath   = "did"
	increment    = 1440
	user         = "issuernode"
)

var (
	// DidNotFound error
	DidNotFound = errors.New("did not found in vault")
	// VaultConnErr error
	VaultConnErr = errors.New("vault connection error")
)

// HTTPClientTimeout http client timeout TODO: move to config
const HTTPClientTimeout = 10 * time.Second

// Config vault configuration
// If UserPassAuthEnabled is true, then vault client will be created with userpass auth and Pass must be provided
type Config struct {
	Address             string
	UserPassAuthEnabled bool
	Token               string
	Pass                string
	TLSEnabled          bool
	CertPath            string
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
		vaultCli, _, err = newVaultClientWithUserPassAuth(ctx, cfg)
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
		vaultCli, err = newVaultClientWithToken(cfg)
		if err != nil {
			log.Error(ctx, "cannot init vault client: ", "err", err)
			return nil, err
		}
	}

	return vaultCli, nil
}

// newVaultClientWithToken checks vault configuration and creates new vault client
func newVaultClientWithToken(cfg Config) (*vault.Client, error) {
	if cfg.Address == "" {
		return nil, errors.New("vault address is not specified")
	}
	if cfg.Address == "" {
		return nil, errors.New("vault access token is not specified")
	}

	config := vault.DefaultConfig()
	if cfg.TLSEnabled {
		err := config.ConfigureTLS(&vault.TLSConfig{
			CACert: cfg.CertPath,
		})
		if err != nil {
			return nil, err
		}
	}
	config.Address = cfg.Address
	config.HttpClient.Timeout = HTTPClientTimeout
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.SetToken(cfg.Token)
	return client, nil
}

// newVaultClientWithUserPassAuth checks vault configuration and creates new vault client with userpass auth
func newVaultClientWithUserPassAuth(ctx context.Context, cfg Config) (*vault.Client, *vault.Secret, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Address
	config.HttpClient.Timeout = HTTPClientTimeout

	if cfg.TLSEnabled {
		err := config.ConfigureTLS(&vault.TLSConfig{
			CACert: cfg.CertPath,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	client, err := vault.NewClient(config)
	if err != nil {
		log.Error(ctx, "error creating vault client with userpass auth", "error", err)
		return nil, nil, err
	}

	secret, err := login(ctx, client, user, cfg.Pass)
	if err != nil {
		log.Error(ctx, "error logging in to vault with userpass auth", "error", err)
		return nil, nil, err
	}

	log.Info(ctx, "successfully logged in to vault with userpass auth", "token", secret.Auth.ClientToken)
	return client, secret, nil
}

func login(ctx context.Context, client *vault.Client, user string, pass string) (*vault.Secret, error) {
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

	return secret, nil
}

// RenewToken renews token
func RenewToken(ctx context.Context, client *vault.Client, cfg Config) {
	for {
		vaultLoginResp, err := login(ctx, client, user, cfg.Pass)
		if err != nil {
			log.Error(ctx, "unable to authenticate to Vault: %v", "err", err)
		}
		tokenErr := manageTokenLifecycle(ctx, client, vaultLoginResp)
		if tokenErr != nil {
			log.Error(ctx, "unable to start managing token lifecycle: %v", "err", tokenErr)
		}
	}
}

func manageTokenLifecycle(ctx context.Context, client *vault.Client, token *vault.Secret) error {
	renew := token.Auth.Renewable // You may notice a different top-level field called Renewable. That one is used for dynamic secrets renewal, not token renewal.
	if !renew {
		log.Info(ctx, "Token is not configured to be renewable. Re-attempting login.")
		return nil
	}

	watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{
		Secret:    token,
		Increment: increment,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize new lifetime watcher for renewing auth token: %w", err)
	}

	go watcher.Start()
	defer watcher.Stop()

	for {
		select {
		// `DoneCh` will return if renewal fails, or if the remaining lease
		// duration is under a built-in threshold and either renewing is not
		// extending it or renewing is disabled. In any case, the caller
		// needs to attempt to log in again.
		case err := <-watcher.DoneCh():
			if err != nil {
				log.Error(ctx, "Failed to renew token: %v. Re-attempting login.", "error", err)
				return nil
			}
			// This occurs once the token has reached max TTL.
			log.Info(ctx, "Token can no longer be renewed. Re-attempting login.")
			return nil

		// Successfully completed renewal
		case renewal := <-watcher.RenewCh():
			log.Info(ctx, "Vault token successfully renewed", "renewal", renewal.RenewedAt)
		}
	}
}

// GetDID gets did from vault
func GetDID(ctx context.Context, vaultCli *vault.Client) (string, error) {
	did, err := vaultCli.KVv2(didMountPath).Get(ctx, secretPath)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			log.Error(ctx, "error getting did from vault, access denied", "error", err)
			return "", errors.Join(err, VaultConnErr)
		}
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
