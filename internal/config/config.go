package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/spf13/viper"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

const (
	CIConfigPath = "/home/runner/work/sh-id-platform/sh-id-platform/" // CIConfigPath variable contain the CI configuration path
	ipfsGateway  = "https://cloudflare-ipfs.com"
)

// Configuration holds the project configuration
type Configuration struct {
	ServerUrl                    string
	ServerPort                   int
	NativeProofGenerationEnabled bool
	Database                     Database      `mapstructure:"Database"`
	Cache                        Cache         `mapstructure:"Cache"`
	HTTPBasicAuth                HTTPBasicAuth `mapstructure:"HTTPBasicAuth"`
	KeyStore                     KeyStore      `mapstructure:"KeyStore"`
	Log                          Log           `mapstructure:"Log"`
	Ethereum                     Ethereum      `mapstructure:"Ethereum"`
	Prover                       Prover        `mapstructure:"Prover"`
	Circuit                      Circuit       `mapstructure:"Circuit"`
	PublishingKeyPath            string        `mapstructure:"PublishingKeyPath"`
	OnChainCheckStatusFrequency  time.Duration `mapstructure:"OnChainCheckStatusFrequency"`
	SchemaCache                  *bool         `mapstructure:"SchemaCache"`
	APIUI                        APIUI         `mapstructure:"APIUI"`
	IPFS                         IPFS          `mapstructure:"IPFS"`
	VaultUserPassAuthEnabled     bool
	VaultUserPassAuthPassword    string
	CredentialStatus             CredentialStatus `mapstructure:"CredentialStatus"`
}

// Database has the database configuration
// URL: The database connection string
type Database struct {
	URL string `mapstructure:"Url" tip:"The Datasource name locator"`
}

// Cache configurations
type Cache struct {
	RedisUrl string `mapstructure:"RedisUrl" tip:"The redis url to use as a cache"`
}

// IPFS configurations
type IPFS struct {
	GatewayURL string `mapstructure:"GatewayUrl" tip:"The IPFS gateway url. (e.g. https://ipfs.io)"`
}

// ReverseHashService contains the reverse hash service properties
type ReverseHashService struct {
	URL     string `mapstructure:"Url" tip:"Reverse Hash Service address"`
	Enabled bool   `tip:"Reverse hash service enabled"`
	Mode    string `tip:"Reverse hash service mode (OffChain, OnChain, Mixed, None)"`
}

// Ethereum struct
type Ethereum struct {
	URL                       string        `tip:"Ethereum url"`
	ContractAddress           string        `tip:"Contract Address"`
	DefaultGasLimit           int           `tip:"Default Gas Limit"`
	ConfirmationTimeout       time.Duration `tip:"Confirmation timeout"`
	ConfirmationBlockCount    int64         `tip:"Confirmation block count"`
	ReceiptTimeout            time.Duration `tip:"Receipt timeout"`
	MinGasPrice               int           `tip:"Minimum Gas Price"`
	MaxGasPrice               int           `tip:"The Datasource name locator"`
	RPCResponseTimeout        time.Duration `tip:"RPC Response timeout"`
	WaitReceiptCycleTime      time.Duration `tip:"Wait Receipt Cycle Time"`
	WaitBlockCycleTime        time.Duration `tip:"Wait Block Cycle Time"`
	ResolverPrefix            string        `tip:"blockchain:network e.g polygon:mumbai"`
	InternalTransferAmountWei int64         `tip:"Internal transfer amount in wei"`
	TransferAccountKeyPath    string        `tip:"Transfer account key path"`
}

// Prover struct
type Prover struct {
	ServerURL       string
	ResponseTimeout time.Duration
}

// Circuit struct
type Circuit struct {
	Path string `tip:"Circuit path"`
}

// KeyStore defines the keystore
type KeyStore struct {
	Address              string `tip:"Keystore address"`
	Token                string `tip:"Token"`
	PluginIden3MountPath string `tip:"PluginIden3MountPath"`
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
	Level int `mapstructure:"Level" tip:"Minimum level to log: (-4:Debug, 0:Info, 4:Warning, 8:Error)"`
	Mode  int `mapstructure:"Mode" tip:"Log format (1: JSON, 2:Structured text)"`
}

// HTTPBasicAuth configuration. Some of the endpoints are protected with basic http auth. Here you can set the
// user and password to use.
type HTTPBasicAuth struct {
	User     string `mapstructure:"User" tip:"Basic auth username"`
	Password string `mapstructure:"Password" tip:"Basic auth password"`
}

// APIUI - APIUI backend service configuration.
type APIUI struct {
	ServerPort         int       `mapstructure:"ServerPort" tip:"Server UI API backend port"`
	ServerURL          string    `mapstructure:"ServerUrl" tip:"Server UI API backend url"`
	APIUIAuth          APIUIAuth `mapstructure:"APIUIAuth" tip:"Server UI API backend basic auth credentials"`
	IssuerName         string    `mapstructure:"IssuerName" tip:"Server UI API backend issuer name"`
	IssuerLogo         string    `mapstructure:"IssuerLogo" tip:"Server UI API backend issuer logo (URL)"`
	Issuer             string    `mapstructure:"IssuerDID" tip:"Server UI API backend issuer DID (already created in the issuer node)"`
	IssuerDID          w3c.DID   `mapstructure:"-"`
	SchemaCache        *bool     `mapstructure:"SchemaCache" tip:"Server UI API backend for enabling schema caching"`
	IdentityMethod     string    `mapstructure:"IdentityMethod" tip:"Server UI API backend Identity Method"`
	IdentityBlockchain string    `mapstructure:"IdentityBlockchain" tip:"Server UI API backend Identity Blockchain"`
	IdentityNetwork    string    `mapstructure:"IdentityNetwork" tip:"Server UI API backend Identity Network"`
	KeyType            string    `mapstructure:"KeyType" tip:"Server UI API backend Key Type"`
}

// APIUIAuth configuration. Some of the UI API endpoints are protected with basic http auth. Here you can set the
// user and password to use.
type APIUIAuth struct {
	User     string `mapstructure:"User" tip:"Server UI APIBasic auth username"`
	Password string `mapstructure:"Password" tip:"Server UI API Basic auth password"`
}

// Sanitize perform some basic checks and sanitizations in the configuration.
// Returns true if config is acceptable, error otherwise.
func (c *Configuration) Sanitize(ctx context.Context) error {
	sUrl, err := c.validateServerUrl()
	if err != nil {
		return fmt.Errorf("serverUrl is not a valid URL <%s>: %w", c.ServerUrl, err)
	}
	c.ServerUrl = sUrl
	if c.KeyStore.Token == "" && !c.VaultUserPassAuthEnabled {
		log.Error(ctx, "a vault token must be provided or vault userpass auth must be enabled", "vaultUserPassAuthEnabled", c.VaultUserPassAuthEnabled)
		return fmt.Errorf("a vault token must be provided or vault userpass auth must be enabled")
	}

	err = c.sanitizeCredentialStatus(ctx, c.ServerUrl)
	if err != nil {
		log.Error(ctx, "error sanitizing credential status", "error", err)
		return err
	}

	return nil
}

// SanitizeAPIUI perform some basic checks and sanitizations in the configuration.
// Returns true if config is acceptable, error otherwise.
func (c *Configuration) SanitizeAPIUI(ctx context.Context) (err error) {
	if c.APIUI.ServerPort == 0 {
		return fmt.Errorf("a port for the UI API server must be provided")
	}

	if c.APIUI.ServerURL == "" {
		return fmt.Errorf("the UI API server url must be provided")
	}

	log.Info(ctx, "Checking vault token", "token", c.KeyStore.Token)
	if c.KeyStore.Token == "" && !c.VaultUserPassAuthEnabled {
		log.Error(ctx, "a vault token must be provided or vault userpass auth must be enabled", "vaultUserPassAuthEnabled", c.VaultUserPassAuthEnabled)
		return fmt.Errorf("a vault token must be provided or vault userpass auth must be enabled")
	}

	if c.APIUI.Issuer != "" {
		issuerDID, err := w3c.ParseDID(c.APIUI.Issuer)
		if err != nil {
			log.Error(ctx, "invalid issuer did format", "error", err)
			return fmt.Errorf("invalid issuer did format")
		}
		c.APIUI.IssuerDID = *issuerDID
	} else {
		log.Info(ctx, "Issuer DID not provided in configuration file")
	}

	err = c.sanitizeCredentialStatus(ctx, c.APIUI.ServerURL)
	if err != nil {
		log.Error(ctx, "error sanitizing credential status", "error", err)
		return err
	}
	return nil
}

func (c *Configuration) sanitizeCredentialStatus(_ context.Context, host string) error {
	if c.CredentialStatus.RHSMode != onChain && c.CredentialStatus.RHSMode != offChain && c.CredentialStatus.RHSMode != none {
		return fmt.Errorf("ISSUER_CREDENTIAL_STATUS_RHS_MODE value is not valid")
	}

	if c.CredentialStatus.RHSMode == none {
		c.CredentialStatus.DirectStatus.URL = host
		c.CredentialStatus.CredentialStatusType = sparseMerkleTreeProof
	}

	if c.CredentialStatus.RHSMode == offChain {
		if c.CredentialStatus.RHS.URL == "" {
			return fmt.Errorf("ISSUER_CREDENTIAL_STATUS_RHS_URL value is missing")
		}
		c.CredentialStatus.CredentialStatusType = iden3ReverseSparseMerkleTreeProof
		c.CredentialStatus.DirectStatus.URL = host
	}

	if c.CredentialStatus.RHSMode == onChain {
		if c.CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract == "" {
			return fmt.Errorf("ISSUER_CREDENTIAL_STATUS_ONCHAIN_TREE_STORE_SUPPORTED_CONTRACT value is missing")
		}

		if c.CredentialStatus.OnchainTreeStore.PublishingKeyPath == "" {
			return fmt.Errorf("ISSUER_CREDENTIAL_STATUS_PUBLISHING_KEY_PATH value is missing")
		}

		if c.CredentialStatus.OnchainTreeStore.ChainID == "" {
			return fmt.Errorf("ISSUER_CREDENTIAL_STATUS_RHS_CHAIN_ID value is missing")
		}
		c.CredentialStatus.CredentialStatusType = iden3OnchainSparseMerkleTreeProof2023
	}
	return nil
}

// CheckDID checks if the issuer did is provided in the configuration file. If not, it tries to get it from vault.
func CheckDID(ctx context.Context, cfg *Configuration, vaultCli *api.Client) error {
	log.Info(ctx, "Checking issuer did value", "did", cfg.APIUI.Issuer)
	if cfg.APIUI.Issuer == "" {
		log.Info(ctx, "Issuer DID not provided in configuration file. Getting it from vault")
		var err error
		cfg.APIUI.Issuer, err = providers.GetDID(ctx, vaultCli)
		if err != nil {
			log.Error(ctx, "cannot get issuer did from vault", "error", err)
			return err
		}
		log.Info(ctx, "Issuer Did loaded from vault", "did", cfg.APIUI.Issuer)
		issuerDID, err := w3c.ParseDID(cfg.APIUI.Issuer)
		if err != nil {
			log.Error(ctx, "invalid issuer did format", "error", err)
			return err
		}
		cfg.APIUI.IssuerDID = *issuerDID
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

// Load loads the configuration from a file
func Load(fileName string) (*Configuration, error) {
	bindEnv()
	pathFlag := viper.GetString("config")
	if _, err := os.Stat(pathFlag); err == nil {
		ext := filepath.Ext(pathFlag)
		if len(ext) > 1 {
			ext = ext[1:]
		}
		name := strings.Split(filepath.Base(pathFlag), ".")[0]
		viper.AddConfigPath(".")
		viper.SetConfigName(name)
		viper.SetConfigType(ext)
	} else {
		// Read default config file.
		viper.AddConfigPath(getWorkingDirectory())
		viper.AddConfigPath(CIConfigPath)
		viper.SetConfigType("toml")
		if fileName == "" {
			viper.SetConfigName("config")
		} else {
			viper.SetConfigName(fileName)
		}
	}
	// const defDBPort = 5432
	config := &Configuration{
		// ServerPort: defDBPort,
		Database: Database{},
		Log: Log{
			Level: log.LevelDebug,
			Mode:  log.OutputText,
		},
	}
	ctx := context.Background()
	if err := viper.ReadInConfig(); err != nil {
		log.Info(ctx, "missing toml config file. Fallback to env vars", "err", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		log.Error(ctx, "error unmarshalling configuration", "err", err)
	}
	checkEnvVars(ctx, config)
	return config, nil
}

// VaultTest returns the vault configuration to be used in tests.
// The vault token is obtained from environment vars.
// If there is no env var, it will try to parse the init.out file
// created by local docker image provided for TESTING purposes.
func VaultTest() KeyStore {
	return KeyStore{
		Address:              "http://localhost:8200",
		Token:                lookupVaultTestToken(),
		PluginIden3MountPath: "iden3",
	}
}

func lookupVaultTestToken() string {
	var err error
	token, ok := os.LookupEnv("VAULT_TEST_TOKEN")
	if !ok {
		token, err = lookupVaultTokenFromFile("infrastructure/local/.vault/data/init.out")
		if err != nil {
			return ""
		}
	}
	return token
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

func bindEnv() {
	viper.SetEnvPrefix("ISSUER")
	_ = viper.BindEnv("ServerUrl", "ISSUER_SERVER_URL")
	_ = viper.BindEnv("ServerPort", "ISSUER_SERVER_PORT")
	_ = viper.BindEnv("NativeProofGenerationEnabled", "ISSUER_NATIVE_PROOF_GENERATION_ENABLED")
	_ = viper.BindEnv("PublishingKeyPath", "ISSUER_PUBLISH_KEY_PATH")
	_ = viper.BindEnv("OnChainCheckStatusFrequency", "ISSUER_ONCHAIN_CHECK_STATUS_FREQUENCY")

	_ = viper.BindEnv("Database.URL", "ISSUER_DATABASE_URL")

	_ = viper.BindEnv("Log.Level", "ISSUER_LOG_LEVEL")
	_ = viper.BindEnv("Log.Mode", "ISSUER_LOG_MODE")

	_ = viper.BindEnv("HTTPBasicAuth.User", "ISSUER_API_AUTH_USER")
	_ = viper.BindEnv("HTTPBasicAuth.Password", "ISSUER_API_AUTH_PASSWORD")

	_ = viper.BindEnv("IPFS.GatewayUrl", "ISSUER_IPFS_GATEWAY_URL")

	_ = viper.BindEnv("KeyStore.Address", "ISSUER_KEY_STORE_ADDRESS")
	_ = viper.BindEnv("KeyStore.Token", "ISSUER_KEY_STORE_TOKEN")
	_ = viper.BindEnv("KeyStore.PluginIden3MountPath", "ISSUER_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH")

	_ = viper.BindEnv("Ethereum.URL", "ISSUER_ETHEREUM_URL")
	_ = viper.BindEnv("Ethereum.ContractAddress", "ISSUER_ETHEREUM_CONTRACT_ADDRESS")
	_ = viper.BindEnv("Ethereum.DefaultGasLimit", "ISSUER_ETHEREUM_DEFAULT_GAS_LIMIT")
	_ = viper.BindEnv("Ethereum.ConfirmationTimeout", "ISSUER_ETHEREUM_CONFIRMATION_TIME_OUT")
	_ = viper.BindEnv("Ethereum.ConfirmationBlockCount", "ISSUER_ETHEREUM_CONFIRMATION_BLOCK_COUNT")
	_ = viper.BindEnv("Ethereum.ReceiptTimeout", "ISSUER_ETHEREUM_RECEIPT_TIMEOUT")
	_ = viper.BindEnv("Ethereum.MinGasPrice", "ISSUER_ETHEREUM_MIN_GAS_PRICE")
	_ = viper.BindEnv("Ethereum.MaxGasPrice", "ISSUER_ETHEREUM_MAX_GAS_PRICE")
	_ = viper.BindEnv("Ethereum.RPCResponseTimeout", "ISSUER_ETHEREUM_RPC_RESPONSE_TIMEOUT")
	_ = viper.BindEnv("Ethereum.WaitReceiptCycleTime", "ISSUER_ETHEREUM_WAIT_RECEIPT_CYCLE_TIME")
	_ = viper.BindEnv("Ethereum.WaitBlockCycleTime", "ISSUER_ETHEREUM_WAIT_BLOCK_CYCLE_TIME")
	_ = viper.BindEnv("Ethereum.ResolverPrefix", "ISSUER_ETHEREUM_RESOLVER_PREFIX")
	_ = viper.BindEnv("Ethereum.InternalTransferAmountWei", "ISSUER_INTERNAL_TRANSFER_AMOUNT_WEI")
	_ = viper.BindEnv("Ethereum.TransferAccountKeyPath", "ISSUER_ETHEREUM_TRANSFER_ACCOUNT_KEY_PATH")

	_ = viper.BindEnv("CredentialStatus.DirectStatus.URL", "ISSUER_CREDENTIAL_STATUS_DIRECT_URL")
	_ = viper.BindEnv("CredentialStatus.RHS.URL", "ISSUER_CREDENTIAL_STATUS_RHS_URL")
	_ = viper.BindEnv("CredentialStatus.OnchainTreeStore.SupportedTreeStoreContract", "ISSUER_CREDENTIAL_STATUS_ONCHAIN_TREE_STORE_SUPPORTED_CONTRACT")
	_ = viper.BindEnv("CredentialStatus.OnchainTreeStore.PublishingKeyPath", "ISSUER_CREDENTIAL_STATUS_PUBLISHING_KEY_PATH")
	_ = viper.BindEnv("CredentialStatus.OnchainTreeStore.ChainID", "ISSUER_CREDENTIAL_STATUS_RHS_CHAIN_ID")
	_ = viper.BindEnv("CredentialStatus.RHSMode", "ISSUER_CREDENTIAL_STATUS_RHS_MODE")
	_ = viper.BindEnv("CredentialStatus.CredentialStatusType", "ISSUER_CREDENTIAL_STATUS_RHS_STATUS_TYPE")

	_ = viper.BindEnv("Prover.ServerURL", "ISSUER_PROVER_SERVER_URL")
	_ = viper.BindEnv("Prover.ResponseTimeout", "ISSUER_PROVER_TIMEOUT")

	_ = viper.BindEnv("Circuit.Path", "ISSUER_CIRCUIT_PATH")

	_ = viper.BindEnv("Cache.RedisUrl", "ISSUER_REDIS_URL")
	_ = viper.BindEnv("SchemaCache", "ISSUER_SCHEMA_CACHE")

	_ = viper.BindEnv("VaultUserPassAuthEnabled", "ISSUER_VAULT_USERPASS_AUTH_ENABLED")
	_ = viper.BindEnv("VaultUserPassAuthPassword", "ISSUER_VAULT_USERPASS_AUTH_PASSWORD")

	_ = viper.BindEnv("APIUI.ServerPort", "ISSUER_API_UI_SERVER_PORT")
	_ = viper.BindEnv("APIUI.ServerURL", "ISSUER_API_UI_SERVER_URL")
	_ = viper.BindEnv("APIUI.APIUIAuth.User", "ISSUER_API_UI_AUTH_USER")
	_ = viper.BindEnv("APIUI.APIUIAuth.Password", "ISSUER_API_UI_AUTH_PASSWORD")
	_ = viper.BindEnv("APIUI.IssuerName", "ISSUER_API_UI_ISSUER_NAME")
	_ = viper.BindEnv("APIUI.IssuerLogo", "ISSUER_API_UI_ISSUER_LOGO")
	_ = viper.BindEnv("APIUI.IssuerDID", "ISSUER_API_UI_ISSUER_DID")
	_ = viper.BindEnv("APIUI.SchemaCache", "ISSUER_API_UI_SCHEMA_CACHE")
	_ = viper.BindEnv("APIUI.IdentityMethod", "ISSUER_API_IDENTITY_METHOD")
	_ = viper.BindEnv("APIUI.IdentityBlockchain", "ISSUER_API_IDENTITY_BLOCKCHAIN")
	_ = viper.BindEnv("APIUI.IdentityNetwork", "ISSUER_API_IDENTITY_NETWORK")
	_ = viper.BindEnv("APIUI.KeyType", "ISSUER_API_KEY_TYPE")

	viper.AutomaticEnv()
}

// nolint:gocyclo
func checkEnvVars(ctx context.Context, cfg *Configuration) {
	if cfg.IPFS.GatewayURL == "" {
		log.Warn(ctx, "ISSUER_IPFS_GATEWAY_URL value is missing, using default value: "+ipfsGateway)
		cfg.IPFS.GatewayURL = ipfsGateway
	}

	if cfg.ServerUrl == "" {
		log.Info(ctx, "ISSUER_SERVER_URL value is missing")
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

	if cfg.Ethereum.URL == "" {
		log.Info(ctx, "ISSUER_ETHEREUM_URL value is missing")
	}

	if cfg.Ethereum.ContractAddress == "" {
		log.Info(ctx, "ISSUER_ETHEREUM_CONTRACT_ADDRESS value is missing")
	}

	if cfg.Ethereum.URL == "" {
		log.Info(ctx, "ISSUER_ETHEREUM_URL value is missing")
	}

	if cfg.Ethereum.DefaultGasLimit == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_DEFAULT_GAS_LIMIT value is missing")
	}

	if cfg.Ethereum.ConfirmationTimeout == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_CONFIRMATION_TIME_OUT value is missing")
	}

	if cfg.Ethereum.ConfirmationBlockCount == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_CONFIRMATION_BLOCK_COUNT value is missing")
	}

	if cfg.Ethereum.ReceiptTimeout == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_RECEIPT_TIMEOUT value is missing")
	}

	if cfg.Ethereum.MaxGasPrice == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_MAX_GAS_PRICE value is missing or is 0")
	}

	if cfg.Ethereum.RPCResponseTimeout == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_RPC_RESPONSE_TIMEOUT value is missing")
	}

	if cfg.Ethereum.WaitReceiptCycleTime == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_WAIT_RECEIPT_CYCLE_TIME value is missing")
	}

	if cfg.Ethereum.WaitBlockCycleTime == 0 {
		log.Info(ctx, "ISSUER_ETHEREUM_WAIT_BLOCK_CYCLE_TIME value is missing")
	}

	if cfg.Ethereum.ResolverPrefix == "" {
		log.Info(ctx, "ISSUER_ETHEREUM_RESOLVER_PREFIX value is missing")
	}

	if cfg.Prover.ServerURL == "" {
		log.Info(ctx, "ISSUER_PROVER_SERVER_URL value is missing")
	}

	if cfg.Prover.ResponseTimeout == 0 {
		log.Info(ctx, "ISSUER_PROVER_TIMEOUT value is missing")
	}

	if cfg.Circuit.Path == "" {
		log.Info(ctx, "ISSUER_CIRCUIT_PATH value is missing")
	}

	if cfg.Cache.RedisUrl == "" {
		log.Info(ctx, "ISSUER_REDIS_URL value is missing")
	}

	if cfg.SchemaCache == nil {
		log.Info(ctx, "ISSUER_SCHEMA_CACHE is missing and the server set up it as false")
		cfg.SchemaCache = common.ToPointer(false)
	}

	if cfg.CredentialStatus.RHSMode == "" {
		log.Info(ctx, "ISSUER_CREDENTIAL_STATUS_RHS_MODE value is missing and the server set up it as None")
		cfg.CredentialStatus.RHSMode = "None"
	}

	if cfg.APIUI.KeyType == "" {
		log.Info(ctx, "ISSUER_KEY_TYPE is missing and the server set up it as BJJ")
		cfg.APIUI.KeyType = "BJJ"
	}

	if cfg.APIUI.ServerPort == 0 {
		log.Info(ctx, "ISSUER_API_UI_SERVER_PORT value is missing")
	}

	if cfg.APIUI.ServerURL == "" {
		log.Info(ctx, "ISSUER_API_UI_SERVER_URL value is missing")
	}

	if cfg.APIUI.APIUIAuth.User == "" {
		log.Info(ctx, "ISSUER_API_UI_AUTH_USER value is missing")
	}

	if cfg.APIUI.APIUIAuth.Password == "" {
		log.Info(ctx, "ISSUER_API_UI_AUTH_PASSWORD value is missing")
	}

	if cfg.APIUI.IssuerName == "" {
		log.Info(ctx, "ISSUER_API_UI_ISSUER_NAME value is missing")
	}

	if cfg.APIUI.Issuer == "" {
		log.Info(ctx, "ISSUER_API_UI_ISSUER_DID value is missing")
	}

	if cfg.APIUI.SchemaCache == nil {
		log.Info(ctx, "ISSUER_API_UI_SCHEMA_CACHE is missing and the server set up it as false")
		cfg.APIUI.SchemaCache = common.ToPointer(false)
	}

	if cfg.APIUI.IdentityMethod == "" {
		log.Info(ctx, "ISSUER_API_IDENTITY_METHOD value is missing and the server set up it as polygonid")
		cfg.APIUI.IdentityMethod = "polygonid"
	}

	if cfg.APIUI.IdentityBlockchain == "" {
		log.Info(ctx, "ISSUER_API_IDENTITY_BLOCKCHAIN value is missing and the server set up it as polygon")
		cfg.APIUI.IdentityBlockchain = "polygon"
	}

	if cfg.APIUI.IdentityNetwork == "" {
		log.Info(ctx, "ISSUER_API_IDENTITY_NETWORK value is missing and the server set up it as mumbai")
		cfg.APIUI.IdentityNetwork = "mumbai"
	}
}

func getWorkingDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..") + "/"
}
