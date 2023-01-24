package config

import (
	"fmt"
	"math/big"
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
	KeyStore                     KeyStore           `mapstructure:"KeyStore"`
	Runtime                      Runtime            `mapstructure:"Runtime"`
	ReverseHashService           ReverseHashService `mapstructure:"ReverseHashService"`
	Ethereum                     Ethereum           `mapstructure:"Ethereum"`
	Prover                       Prover             `mapstructure:"Prover"`
	Circuit                      Circuit            `mapstructure:"Circuit"`
	PublishingKeyPath            string             `mapstructure:"PublishingKeyPath"`
	OnChainPublishStateFrecuency string             `mapstructure:"OnChainPublishStateFrecuency"`
	OnChainCheckStatusFrecuency  string             `mapstructure:"OnChainCheckStatusFrecuency"`
}

// Database has the database configuration
// URL: The database connection string
type Database struct {
	URL string
}

// ReverseHashService contains the reverse hash service properties
type ReverseHashService struct {
	URL     string
	Enabled bool
}

// Ethereum struct
type Ethereum struct {
	URL                    string
	ContractAddress        string
	DefaultGasLimit        int
	ConfirmationTimeout    time.Duration
	ConfirmationBlockCount int64
	ReceiptTimeout         time.Duration
	MinGasPrice            *big.Int
	MaxGasPrice            *big.Int
	RPCResponseTimeout     time.Duration
	WaitReceiptCycleTime   time.Duration
	WaitBlockCycleTime     time.Duration
}

// Prover struct
type Prover struct {
	ServerURL       string
	ResponseTimeout time.Duration
}

// Circuit struct
type Circuit struct {
	Path string
}

// KeyStore defines the keystore
type KeyStore struct {
	Address              string
	Token                string
	PluginIden3MountPath string
}

// Runtime holds runtime configurations
//
// LogLevel: The minimum log level to show on logs. Values can be
//
//	 -4: Debug
//		0: Info
//		4: Warning
//		8: Error
//	 The default log level is debug
//
// LogMode: Log mode is the format of the log. It can be text or json
// 1: JSON
// 2: Text
// The default log formal is JSON
type Runtime struct {
	LogLevel int `mapstructure:"LogLevel"`
	LogMode  int `mapstructure:"LogMode"`
}

// Load loads the configuraion from a file
func Load(fileName string) (*Configuration, error) {
	//if err := getFlags(); err != nil {
	//	return nil, err
	//}
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
		Runtime: Runtime{
			LogLevel: log.LevelDebug,
			LogMode:  log.OutputText,
		},
	}

	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(config); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

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

//func getFlags() error {
//	pflag.StringP("config", "c", "", "Specify the configuration file location.")
//	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
//	pflag.Parse()
//
//	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
//		log.Error(context.Background(), "parsing config flags", err)
//		return err
//	}
//	return nil
//}

func bindEnv() {
	viper.SetEnvPrefix("SH_ID_PLATFORM")
	_ = viper.BindEnv("ServerUrl", "SH_ID_PLATFORM_SERVER_URL")
	_ = viper.BindEnv("ServerPort", "SH_ID_PLATFORM_SERVER_PORT")
	_ = viper.BindEnv("Database.URL", "SH_ID_PLATFORM_DATABASE_URL")
	viper.AutomaticEnv()
}

func getWorkingDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..") + "/"
}
