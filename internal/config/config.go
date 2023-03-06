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

	"github.com/spf13/viper"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// CIConfigPath variable contain the CI configuration path
const CIConfigPath = "/home/runner/work/sh-id-platform/sh-id-platform/"

// Configuration holds the project configuration
type Configuration struct {
	ServerUrl                    string
	ServerPort                   int
	NativeProofGenerationEnabled bool
	Database                     Database           `mapstructure:"Database"`
	Cache                        Cache              `mapstructure:"Cache"`
	HTTPBasicAuth                HTTPBasicAuth      `mapstructure:"HTTPBasicAuth"`
	KeyStore                     KeyStore           `mapstructure:"KeyStore"`
	Log                          Log                `mapstructure:"Log"`
	ReverseHashService           ReverseHashService `mapstructure:"ReverseHashService"`
	Ethereum                     Ethereum           `mapstructure:"Ethereum"`
	Prover                       Prover             `mapstructure:"Prover"`
	Circuit                      Circuit            `mapstructure:"Circuit"`
	PublishingKeyPath            string             `mapstructure:"PublishingKeyPath"`
	OnChainCheckStatusFrecuency  time.Duration      `mapstructure:"OnChainCheckStatusFrecuency"`

	ApiUI ApiUI `mapstructure:"ApiUI"`
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

// ReverseHashService contains the reverse hash service properties
type ReverseHashService struct {
	URL     string `mapstructure:"Url" tip:"Reverse Hash Service address"`
	Enabled bool   `tip:"Reverse hash service enabled"`
}

// Ethereum struct
type Ethereum struct {
	URL                    string        `tip:"Ethereum url"`
	ContractAddress        string        `tip:"Contract Address"`
	DefaultGasLimit        int           `tip:"Default Gas Limit"`
	ConfirmationTimeout    time.Duration `tip:"Confirmation timeout"`
	ConfirmationBlockCount int64         `tip:"Confirmation block count"`
	ReceiptTimeout         time.Duration `tip:"Receipt timeout"`
	MinGasPrice            int           `tip:"Minimum Gas Price"`
	MaxGasPrice            int           `tip:"The Datasource name locator"`
	RPCResponseTimeout     time.Duration `tip:"RPC Response timeout"`
	WaitReceiptCycleTime   time.Duration `tip:"Wait Receipt Cycle Time"`
	WaitBlockCycleTime     time.Duration `tip:"Wait Block Cycle Time"`
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

// ApiUI - ApiUI backend service configuration.
type ApiUI struct {
	ServerPort    int           `mapstructure:"ServerPort" tip:"Server admin backend port"`
	HTTPAPIUIAuth HTTPAPIUIAuth `mapstructure:"HTTPAPIUIAuth" tip:"Server api backend basic auth credentials"`
	IssuerName    string        `mapstructure:"IssuerName" tip:"Server admin backend issuer name"`
	IssuerLogo    string        `mapstructure:"IssuerLogo" tip:"Server admin backend issuer logo (url)"`
	IssuerDID     string        `mapstructure:"IssuerDID" tip:"Server admin backend issuer DID (already created in the issuer node)"`
}

// HTTPAPIUIAuth configuration. Some of the admin endpoints are protected with basic http auth. Here you can set the
// user and password to use.
type HTTPAPIUIAuth struct {
	User     string `mapstructure:"User" tip:"Basic auth username"`
	Password string `mapstructure:"Password" tip:"Basic auth password"`
}

// Sanitize perform some basic checks and sanitizations in the configuration.
// Returns true if config is acceptable, error otherwise.
func (c *Configuration) Sanitize() error {
	sUrl, err := c.validateServerUrl()
	if err != nil {
		return fmt.Errorf("serverUrl is not a valid url <%s>: %w", c.ServerUrl, err)
	}
	c.ServerUrl = sUrl

	return nil
}

// SanitizeAdmin perform some basic checks and sanitizations in the configuration.
// Returns true if config is acceptable, error otherwise.
func (c *Configuration) SanitizeAdmin() error {

	if c.ApiUI.ServerPort == 0 {
		return fmt.Errorf("the api ui server port should be provided")
	}

	if c.ApiUI.IssuerLogo == "" {
		c.ApiUI.IssuerLogo = "http://no-logo.com"
	}

	if c.ApiUI.IssuerDID == "" {
		return fmt.Errorf("the Issuer DID value is empty and you must to provide one")
	}

	return nil
}

func (c *Configuration) validateServerUrl() (string, error) {
	sUrl, err := url.ParseRequestURI(c.ServerUrl)
	if err != nil {
		return c.ServerUrl, err
	}
	if sUrl.Scheme == "" {
		return c.ServerUrl, fmt.Errorf("server url must be an absolute url")
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
			viper.SetConfigName("config1")
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
	_ = viper.ReadInConfig()
	_ = viper.Unmarshal(config)
	checkEnvVars(config)
	return config, nil
}

// VaultTest returns the vault configuration to be used in tests.
// The vault token is obtained from environment vars.
// If there is not env var, it will try to parse the init.out file
// created by local docker image provided for TESTING purposes.
func VaultTest() KeyStore {
	return KeyStore{
		Address:              "http://localhost:8200",
		Token:                lookupVaultToken(),
		PluginIden3MountPath: "iden3",
	}
}

func lookupVaultToken() string {
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
	viper.SetEnvPrefix("SH_ID_PLATFORM")
	_ = viper.BindEnv("ServerUrl", "SH_ID_PLATFORM_SERVER_URL")
	_ = viper.BindEnv("ServerPort", "SH_ID_PLATFORM_SERVER_PORT")
	_ = viper.BindEnv("NativeProofGenerationEnabled", "SH_ID_PLATFORM_NATIVE_PROOF_GENERATION_ENABLED")
	_ = viper.BindEnv("PublishingKeyPath", "SH_ID_PLATFORM_PUBLISH_KEY_PATH")
	_ = viper.BindEnv("OnChainCheckStatusFrecuency", "SH_ID_PLATFORM_ONCHAIN_CHECK_STATUS_FRECUENCY")

	_ = viper.BindEnv("Database.URL", "SH_ID_PLATFORM_DATABASE_URL")

	_ = viper.BindEnv("Log.Level", "SH_ID_PLATFORM_LOG_LEVEL")
	_ = viper.BindEnv("Log.Mode", "SH_ID_PLATFORM_LOG_MODE")

	_ = viper.BindEnv("HTTPBasicAuth.User", "SH_ID_PLATFORM_HTTPBASICAUTH_USER")
	_ = viper.BindEnv("HTTPBasicAuth.Password", "SH_ID_PLATFORM_HTTPBASICAUTH_PASSWORD")

	_ = viper.BindEnv("KeyStore.Address", "SH_ID_PLATFORM_KEY_STORE_ADDRESS")
	_ = viper.BindEnv("KeyStore.Token", "SH_ID_PLATFORM_KEY_STORE_TOKEN")
	_ = viper.BindEnv("KeyStore.PluginIden3MountPath", "SH_ID_PLATFORM_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH")

	_ = viper.BindEnv("ReverseHashService.URL", "SH_ID_PLATFORM_REVERSE_HASH_SERVICE_URL")
	_ = viper.BindEnv("ReverseHashService.Enabled", "SH_ID_PLATFORM_REVERSE_HASH_SERVICE_ENABLED")

	_ = viper.BindEnv("Ethereum.URL", "SH_ID_PLATFORM_ETHEREUM_URL")
	_ = viper.BindEnv("Ethereum.ContractAddress", "SH_ID_PLATFORM_ETHEREUM_CONTRACT_ADDRESS")
	_ = viper.BindEnv("Ethereum.DefaultGasLimit", "SH_ID_PLATFORM_ETHEREUM_DEFAULT_GAS_LIMIT")
	_ = viper.BindEnv("Ethereum.ConfirmationTimeout", "SH_ID_PLATFORM_ETHEREUM_CONFIRMATION_TIME_OUT")
	_ = viper.BindEnv("Ethereum.ConfirmationBlockCount", "SH_ID_PLATFORM_ETHEREUM_CONFIRMATION_BLOCK_COUNT")
	_ = viper.BindEnv("Ethereum.ReceiptTimeout", "SH_ID_PLATFORM_ETHEREUM_RECEIPT_TIMEOUT")
	_ = viper.BindEnv("Ethereum.MinGasPrice", "SH_ID_PLATFORM_ETHEREUM_MIN_GAS_PRICE")
	_ = viper.BindEnv("Ethereum.MaxGasPrice", "SH_ID_PLATFORM_ETHEREUM_MAX_GAS_PRICE")
	_ = viper.BindEnv("Ethereum.RPCResponseTimeout", "SH_ID_PLATFORM_ETHEREUM_RPC_RESPONSE_TIMEOUT")
	_ = viper.BindEnv("Ethereum.WaitReceiptCycleTime", "SH_ID_PLATFORM_ETHEREUM_WAIT_RECEIPT_CYCLE_TIME")
	_ = viper.BindEnv("Ethereum.WaitBlockCycleTime", "SH_ID_PLATFORM_ETHEREUM_WAIT_BLOCK_CYCLE_TIME")

	_ = viper.BindEnv("Prover.ServerURL", "SH_ID_PLATFORM_PROVER_SERVER_URL")
	_ = viper.BindEnv("Prover.ResponseTimeout", "SH_ID_PLATFORM_PROVER_TIMEOUT")

	_ = viper.BindEnv("Circuit.Path", "SH_ID_PLATFORM_CIRCUIT_PATH")

	_ = viper.BindEnv("Cache.RedisUrl", "SH_ID_PLATFORM_REDIS_URL")

	_ = viper.BindEnv("ApiUI.ServerPort", "SH_ID_PLATFORM_API_UI_SERVER_PORT")
	_ = viper.BindEnv("ApiUI.HTTPAPIUIAuth.User", "SH_ID_PLATFORM_API_UI_AUTH_USER")
	_ = viper.BindEnv("ApiUI.HTTPAPIUIAuth.Password", "SH_ID_PLATFORM_API_UI_AUTH_PASSWORD")
	_ = viper.BindEnv("ApiUI.IssuerName", "SH_ID_PLATFORM_API_UI_ISSUER_NAME")
	_ = viper.BindEnv("ApiUI.IssuerLogo", "SH_ID_PLATFORM_API_UI_ISSUER_LOGO")
	_ = viper.BindEnv("ApiUI.IssuerDID", "SH_ID_PLATFORM_API_UI_ISSUER_DID")

	viper.AutomaticEnv()
}

func checkEnvVars(cfg *Configuration) {
	if cfg.ServerUrl == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_SERVER_URL value is missing")
	}

	if cfg.ServerPort == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_SERVER_PORT value is missing")
	}

	if cfg.PublishingKeyPath == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_PUBLISH_KEY_PATH value is missing")
	}

	if cfg.OnChainCheckStatusFrecuency == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ONCHAIN_CHECK_STATUS_FRECUENCY value is missing")
	}

	if cfg.Database.URL == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_DATABASE_URL value is missing")
	}

	if cfg.HTTPBasicAuth.User == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_HTTPBASICAUTH_USER value is missing")
	}

	if cfg.HTTPBasicAuth.Password == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_HTTPBASICAUTH_PASSWORD value is missing")
	}

	if cfg.KeyStore.Address == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_KEY_STORE_ADDRESS value is missing")
	}

	if cfg.KeyStore.Token == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_KEY_STORE_TOKEN value is missing")
	}

	if cfg.KeyStore.PluginIden3MountPath == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_KEY_STORE_PLUGIN_IDEN3_MOUNT_PATH value is missing")
	}

	if cfg.Ethereum.URL == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_URL value is missing")
	}

	if cfg.Ethereum.ContractAddress == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_CONTRACT_ADDRESS value is missing")
	}

	if cfg.Ethereum.URL == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_URL value is missing")
	}

	if cfg.Ethereum.DefaultGasLimit == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_DEFAULT_GAS_LIMIT value is missing")
	}

	if cfg.Ethereum.ConfirmationTimeout == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_CONFIRMATION_TIME_OUT value is missing")
	}

	if cfg.Ethereum.ConfirmationBlockCount == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_CONFIRMATION_BLOCK_COUNT value is missing")
	}

	if cfg.Ethereum.ReceiptTimeout == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_RECEIPT_TIMEOUT value is missing")
	}

	if cfg.Ethereum.MinGasPrice == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_MIN_GAS_PRICE value is missing or is 0")
	}

	if cfg.Ethereum.MaxGasPrice == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_MAX_GAS_PRICE value is missing or is 0")
	}

	if cfg.Ethereum.RPCResponseTimeout == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_RPC_RESPONSE_TIMEOUT value is missing")
	}

	if cfg.Ethereum.WaitReceiptCycleTime == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_WAIT_RECEIPT_CYCLE_TIME value is missing")
	}

	if cfg.Ethereum.WaitBlockCycleTime == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_ETHEREUM_WAIT_BLOCK_CYCLE_TIME value is missing")
	}

	if cfg.Prover.ServerURL == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_PROVER_SERVER_URL value is missing")
	}

	if cfg.Prover.ResponseTimeout == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_PROVER_TIMEOUT value is missing")
	}

	if cfg.Circuit.Path == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_CIRCUIT_PATH value is missing")
	}

	if cfg.Cache.RedisUrl == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_REDIS_URL value is missing")
	}

	if cfg.ApiUI.ServerPort == 0 {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_SERVER_PORT value is missing")
	}

	if cfg.ApiUI.HTTPAPIUIAuth.User == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_AUTH_USER value is missing")
	}

	if cfg.ApiUI.HTTPAPIUIAuth.Password == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_AUTH_PASSWORD value is missing")
	}

	if cfg.ApiUI.IssuerName == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_ISSUER_NAME value is missing")
	}

	if cfg.ApiUI.IssuerLogo == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_ISSUER_LOGO value is missing")
	}

	if cfg.ApiUI.IssuerDID == "" {
		log.Info(context.Background(), "SH_ID_PLATFORM_API_UI_ISSUER_DID value is missing")
	}
}

func getWorkingDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..") + "/"
}
