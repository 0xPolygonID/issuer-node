package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	vault "github.com/hashicorp/vault/api"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

const (
	CIConfigPath = "/home/runner/work/sh-id-platform/sh-id-platform/" // CIConfigPath variable contain the CI configuration path
	// LocalStorage is the local storage plugin
	LocalStorage = "localstorage"
	// Vault is the vault plugin
	Vault = "vault"
	// AWSSM is the AWS secret manager provider
	AWSSM = "aws-sm"
	// AWSKMS is the AWS KMS provider
	AWSKMS = "aws-kms"
	// CacheProviderRedis is the redis cache provider
	CacheProviderRedis = "redis"
	// CacheProviderValKey is the valkey cache provider
	CacheProviderValKey = "valkey"

	ipfsGateway = "https://cloudflare-ipfs.com"
)

// Configuration holds the project configuration
type Configuration struct {
	ServerUrl                   string        `env:"ISSUER_SERVER_URL" envDefault:"http://localhost"`
	ServerPort                  int           `env:"ISSUER_SERVER_PORT" envDefault:"3001"`
	PublishingKeyPath           string        `env:"ISSUER_PUBLISH_KEY_PATH" envDefault:"pbkey"`
	SchemaCache                 bool          `env:"ISSUER_SCHEMA_CACHE" envDefault:"false"`
	OnChainCheckStatusFrequency time.Duration `env:"ISSUER_ONCHAIN_CHECK_STATUS_FREQUENCY"`
	NetworkResolverPath         string        `env:"ISSUER_RESOLVER_PATH"`
	NetworkResolverFile         *string       `env:"ISSUER_RESOLVER_FILE"`
	IssuerName                  string        `env:"ISSUER_ISSUER_NAME"`
	IssuerLogo                  string        `env:"ISSUER_ISSUER_LOGO"`
	Database                    Database
	Cache                       Cache
	HTTPBasicAuth               HTTPBasicAuth
	KeyStore                    KeyStore
	Log                         Log
	Ethereum                    Ethereum
	Circuit                     Circuit
	IPFS                        IPFS
	CustomDIDMethods            []CustomDIDMethods `mapstructure:"-"`
	MediaTypeManager            MediaTypeManager
	UniversalLinks              UniversalLinks
	UniversalDIDResolver        UniversalDIDResolver
}

// Database has the database configuration
// URL: The database connection string
type Database struct {
	URL string `env:"ISSUER_DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/issuer?sslmode=disable"`
}

// Cache configurations
type Cache struct {
	Provider string `env:"ISSUER_CACHE_PROVIDER" envDefault:"redis"`
	Url      string `env:"ISSUER_CACHE_URL"`
}

// IPFS configurations
type IPFS struct {
	GatewayURL string `env:"ISSUER_IPFS_GATEWAY_URL" envDefault:"https://cloudflare-ipfs.com"`
}

// Ethereum struct
type Ethereum struct {
	TransferAccountKeyPath string `env:"ISSUER_ETHEREUM_TRANSFER_ACCOUNT_KEY_PATH"`
}

// CustomDIDMethods struct
// Example: ISSUER_CUSTOM_DID_METHODS='[{"blockchain":"linea","network":"testnet","networkFlag":"0b01000001","chainID":59140}]'
type CustomDIDMethods struct {
	Blockchain  string `tip:"Identity blockchain for custom network"`
	Network     string `tip:"Identity network for custom network"`
	NetworkFlag byte   `tip:"Identity network flag for custom network"`
	ChainID     int    `tip:"Chain id for custom network"`
}

// UnmarshalJSON implements the Unmarshal interface for CustomNetwork
func (cn *CustomDIDMethods) UnmarshalJSON(data []byte) error {
	aux := struct {
		Blockchain  string `json:"blockchain"`
		Network     string `json:"network"`
		NetworkFlag string `json:"networkFlag"`
		ChainID     int    `json:"chainId"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.NetworkFlag) != 10 || aux.NetworkFlag[:2] != "0b" {
		return errors.New("invalid NetworkFlag format")
	}
	flag, err := strconv.ParseUint(aux.NetworkFlag[2:], 2, 8)
	if err != nil {
		return err
	}

	cn.Blockchain = aux.Blockchain
	cn.Network = aux.Network
	cn.NetworkFlag = byte(flag)
	cn.ChainID = aux.ChainID

	return nil
}

// Circuit struct
type Circuit struct {
	Path string `env:"ISSUER_CIRCUIT_PATH"`
}

// KeyStore defines the keystore
type KeyStore struct {
	Address                      string `env:"ISSUER_KEY_STORE_ADDRESS"`
	Token                        string `env:"ISSUER_KEY_STORE_TOKEN"`
	PluginIden3MountPath         string `env:"ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH"`
	BJJProvider                  string `env:"ISSUER_KMS_BJJ_PROVIDER"`
	ETHProvider                  string `env:"ISSUER_KMS_ETH_PROVIDER"`
	ProviderLocalStorageFilePath string `env:"ISSUER_KMS_PROVIDER_LOCAL_STORAGE_FILE_PATH"`
	AWSAccessKey                 string `env:"ISSUER_KMS_AWS_ACCESS_KEY"`
	AWSSecretKey                 string `env:"ISSUER_KMS_AWS_SECRET_KEY"`
	AWSRegion                    string `env:"ISSUER_KMS_AWS_REGION"`
	AWSURL                       string `env:"ISSUER_KMS_AWS_URL" envDefault:"http://localstack:4566"`
	VaultUserPassAuthEnabled     bool   `env:"ISSUER_VAULT_USERPASS_AUTH_ENABLED"`
	VaultUserPassAuthPassword    string `env:"ISSUER_VAULT_USERPASS_AUTH_PASSWORD"`
	TLSEnabled                   bool   `env:"ISSUER_VAULT_TLS_ENABLED"`
	CertPath                     string `env:"ISSUER_VAULT_TLS_CERT_PATH"`
}

// UniversalDIDResolver defines the universal DID resolver
type UniversalDIDResolver struct {
	UniversalResolverURL *string `env:"ISSUER_UNIVERSAL_DID_RESOLVER_URL"`
}

// Log holds runtime configurations
//
// Level: The minimum log level to show on logs. Values can be
//
//	 -4: Debug
//		0: Info
//		4: Warning
//		8: Error
//	 The default log level is debug
//
// Mode: Log mode is the format of the log. It can be text or json
// 1: JSON
// 2: Text
// The default log formal is JSON
type Log struct {
	Level int `env:"ISSUER_LOG_LEVEL" envDefault:"0" tip:"Log level (0: Info, 4: Warning, 8: Error)"`
	Mode  int `env:"ISSUER_LOG_MODE" envDefault:"1" tip:"Log mode (1: JSON, 2: Text)"`
}

// HTTPBasicAuth configuration. Some of the endpoints are protected with basic http auth. Here you can set the
// user and password to use.
type HTTPBasicAuth struct {
	User     string `env:"ISSUER_API_AUTH_USER" envDefault:""`
	Password string `env:"ISSUER_API_AUTH_PASSWORD" envDefault:""`
}

// MediaTypeManager enables or disables the media types manager
type MediaTypeManager struct {
	Enabled *bool `env:"ISSUER_MEDIA_TYPE_MANAGER_ENABLED"`
}

// UniversalLinks configuration
type UniversalLinks struct {
	BaseUrl string `env:"ISSUER_UNIVERSAL_LINKS_BASE_URL" envDefault:"https://wallet.privado.id"`
}

// Load loads the configuration from a file
func Load() (*Configuration, error) {
	ctx := context.Background()
	cfg := Configuration{} // 👈 new instance of `Config`
	err := env.Parse(&cfg) // 👈 Parse environment variables into `Config`
	if err != nil {
		return nil, err
	}
	if err := checkEnvVars(ctx, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.sanitize(ctx); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// sanitize perform some basic checks and sanitizations in the configuration.
// Returns true if config is acceptable, error otherwise.
func (c *Configuration) sanitize(ctx context.Context) error {
	sUrl, err := c.validateServerUrl()
	if err != nil {
		return fmt.Errorf("serverUrl is not a valid URL <%s>: %w", c.ServerUrl, err)
	}
	c.ServerUrl = sUrl
	if c.KeyStore.Token == "" && !c.KeyStore.VaultUserPassAuthEnabled {
		log.Error(ctx, "a vault token must be provided or vault userpass auth must be enabled", "vaultUserPassAuthEnabled", c.KeyStore.VaultUserPassAuthEnabled)
		return fmt.Errorf("a vault token must be provided or vault userpass auth must be enabled")
	}

	return nil
}

func (c *Configuration) validateServerUrl() (string, error) {
	sUrl, err := url.ParseRequestURI(c.ServerUrl)
	if err != nil {
		return c.ServerUrl, err
	}
	if sUrl.Scheme == "" {
		return c.ServerUrl, fmt.Errorf("server URL must be an absolute URL")
	}
	sUrl.RawQuery = ""
	return strings.Trim(strings.Trim(sUrl.String(), "/"), "?"), nil
}

// lookupVaultTokenFromFile parses the vault config file looking for the hvs token and returns it
// pathVaultConfig MUST be a relative path starting from the root project folder
// like "infrastructure/local/.vault/data/init.out"
// This function MUST BE only used in tests.
// NEVER share the hvs token in production mode.
func lookupVaultTokenFromFile(pathVaultConfig string) (string, error) {
	r, err := regexp.Compile("hvs.[a-zA-Z0-9]{24}")
	if err != nil {
		return "", fmt.Errorf("wrong regexp: %w", err)
	}
	configFile := getWorkingDirectory() + pathVaultConfig
	content, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}
	matches := r.FindStringSubmatch(string(content))
	if len(matches) != 1 {
		return "", fmt.Errorf("expecting only one match parsing vault config. found %d", len(matches))
	}
	return matches[0], nil
}

// nolint:gocyclo,gocognit
func checkEnvVars(ctx context.Context, cfg *Configuration) error {
	if cfg.IPFS.GatewayURL == "" {
		log.Warn(ctx, "ISSUER_IPFS_GATEWAY_URL value is missing, using default value: "+ipfsGateway)
		cfg.IPFS.GatewayURL = ipfsGateway
	}

	if cfg.ServerUrl == "" {
		log.Error(ctx, "ISSUER_SERVER_URL value is missing")
		return errors.New("ISSUER_SERVER_URL value is missing")
	}

	if cfg.ServerPort == 0 {
		log.Info(ctx, "ISSUER_SERVER_PORT value is missing")
	}

	if cfg.PublishingKeyPath == "" {
		log.Info(ctx, "ISSUER_PUBLISH_KEY_PATH value is missing")
	}

	if cfg.OnChainCheckStatusFrequency == 0 {
		log.Info(ctx, "ISSUER_ONCHAIN_CHECK_STATUS_FREQUENCY value is missing")
	}

	if cfg.Database.URL == "" {
		log.Info(ctx, "ISSUER_DATABASE_URL value is missing")
	}

	if cfg.HTTPBasicAuth.User == "" {
		log.Info(ctx, "ISSUER_API_AUTH_USER value is missing")
	}

	if cfg.HTTPBasicAuth.Password == "" {
		log.Info(ctx, "ISSUER_API_AUTH_PASSWORD value is missing")
	}

	if cfg.KeyStore.Address == "" {
		log.Info(ctx, "ISSUER_KEY_STORE_ADDRESS value is missing")
	}

	if cfg.KeyStore.Token == "" {
		log.Info(ctx, "ISSUER_KEY_STORE_TOKEN value is missing")
	}

	if cfg.KeyStore.PluginIden3MountPath == "" {
		log.Info(ctx, "ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH value is missing")
	}

	if cfg.Circuit.Path == "" {
		log.Info(ctx, "ISSUER_CIRCUIT_PATH value is missing")
	}

	if cfg.Cache.Url == "" {
		log.Error(ctx, "ISSUER_CACHE_URL value is missing")
		return errors.New("ISSUER_CACHE_URL value is missing")
	}

	if cfg.MediaTypeManager.Enabled == nil {
		log.Info(ctx, "ISSUER_MEDIA_TYPE_MANAGER_ENABLED is missing and the server set up it as true")
		cfg.MediaTypeManager.Enabled = common.ToPointer(true)
	}

	if cfg.NetworkResolverPath == "" {
		log.Info(ctx, "ISSUER_RESOLVER_PATH value is missing. Trying to use ISSUER_RESOLVER_FILE")
		if cfg.NetworkResolverFile == nil || *cfg.NetworkResolverFile == "" {
			log.Info(ctx, "ISSUER_RESOLVER_FILE value is missing")
		} else {
			log.Info(ctx, "ISSUER_RESOLVER_FILE value is present")
		}
	}

	if cfg.NetworkResolverPath == "" && (cfg.NetworkResolverFile == nil || *cfg.NetworkResolverFile == "") {
		log.Info(ctx, "ISSUER_RESOLVER_PATH and ISSUER_RESOLVER_FILE value is missing. Using default value: ./resolvers_settings.yaml")
		cfg.NetworkResolverPath = "./resolvers_settings.yaml"
	}

	if cfg.KeyStore.BJJProvider == "" {
		log.Info(ctx, "ISSUER_KMS_BJJ_PLUGIN value is missing, using default value: localstorage")
		cfg.KeyStore.BJJProvider = LocalStorage
	}

	if cfg.KeyStore.ETHProvider == "" {
		log.Info(ctx, "ISSUER_KMS_ETH_PLUGIN value is missing, using default value: localstorage")
		cfg.KeyStore.ETHProvider = LocalStorage
	}

	if (cfg.KeyStore.BJJProvider == LocalStorage || cfg.KeyStore.ETHProvider == LocalStorage) && cfg.KeyStore.ProviderLocalStorageFilePath == "" {
		log.Info(ctx, "ISSUER_KMS_PLUGIN_LOCAL_STORAGE_FOLDER value is missing, using default value: ./localstoragekeys")
		cfg.KeyStore.ProviderLocalStorageFilePath = "./localstoragekeys"
	}

	if cfg.KeyStore.ETHProvider == AWSSM || cfg.KeyStore.ETHProvider == AWSKMS || cfg.KeyStore.BJJProvider == AWSSM {
		if cfg.KeyStore.AWSAccessKey == "" {
			log.Error(ctx, "ISSUER_AWS_KEY_ID value is missing")
			return errors.New("ISSUER_AWS_KEY_ID value is missing")
		}
		if cfg.KeyStore.AWSSecretKey == "" {
			log.Error(ctx, "ISSUER_AWS_SECRET_KEY value is missing")
			return errors.New("ISSUER_AWS_SECRET_KEY value is missing")
		}
		if cfg.KeyStore.AWSRegion == "" {
			log.Error(ctx, "ISSUER_AWS_REGION value is missing")
			return errors.New("ISSUER_AWS_REGION value is missing")
		}
	}

	if cfg.KeyStore.BJJProvider == LocalStorage || cfg.KeyStore.ETHProvider == LocalStorage {
		log.Info(ctx, `
			=====================================================================================================================================================
			IMPORTANT: THIS CONFIGURATION SHOULD NOT BE USED IN PRODUCTIVE ENVIRONMENTS!!!. YOU HAVE CONFIGURED THE ISSUER NODE TO SAVE KEYS IN THE LOCAL STORAGE
			=====================================================================================================================================================
`)
	}

	return nil
}

// KeyStoreConfig initializes the key store
func KeyStoreConfig(ctx context.Context, cfg *Configuration, vaultCfg providers.Config) (*kms.KMS, error) {
	var (
		vaultCli *vault.Client
		vaultErr error
	)
	if cfg.KeyStore.BJJProvider == Vault || cfg.KeyStore.ETHProvider == Vault {
		log.Info(ctx, "using vault key provider")
		vaultCli, vaultErr = providers.VaultClient(ctx, vaultCfg)
		if vaultErr != nil {
			log.Error(ctx, "cannot initialize vault client", "err", vaultErr)
			return nil, vaultErr
		}

		if vaultCfg.UserPassAuthEnabled {
			go providers.RenewToken(ctx, vaultCli, vaultCfg)
		}
	}

	kmsConfig := kms.Config{
		BJJKeyProvider:           kms.ConfigProvider(cfg.KeyStore.BJJProvider),
		ETHKeyProvider:           kms.ConfigProvider(cfg.KeyStore.ETHProvider),
		AWSAccessKey:             cfg.KeyStore.AWSAccessKey,
		AWSSecretKey:             cfg.KeyStore.AWSSecretKey,
		AWSRegion:                cfg.KeyStore.AWSRegion,
		AWSURL:                   cfg.KeyStore.AWSURL,
		LocalStoragePath:         cfg.KeyStore.ProviderLocalStorageFilePath,
		Vault:                    vaultCli,
		PluginIden3MountPath:     cfg.KeyStore.PluginIden3MountPath,
		IssuerETHTransferKeyPath: cfg.Ethereum.TransferAccountKeyPath,
	}

	keyStore, err := kms.OpenWithConfig(ctx, kmsConfig)
	if err != nil {
		log.Error(ctx, "cannot initialize kms", "err", err)
		return nil, err
	}
	return keyStore, nil
}

func getWorkingDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..") + "/"
}
