package config

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/polygonid/sh-id-platform/internal/log"
)

const CIConfigPath = "/home/runner/work/sh-id-platform/sh-id-platform/"

// Configuration holds the project configuration
type Configuration struct {
	ServerUrl          string
	ServerPort         int
	Database           Database           `mapstructure:"Database"`
	KeyStore           KeyStore           `mapstructure:"KeyStore"`
	Runtime            Runtime            `mapstructure:"Runtime"`
	ReverseHashService ReverseHashService `mapstructure:"ReverseHashService"`
}

// Database has the database configuration
// URL: The database connection string
type Database struct {
	URL string
}

type ReverseHashService struct {
	URL     string
	Enabled bool
}

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
	getFlags()
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
	const defDBPort = 5432
	config := &Configuration{
		ServerPort: defDBPort,
		Database:   Database{},
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

func getFlags() {
	pflag.StringP("config", "c", "", "Specify the configuration file location.")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Error(context.Background(), "parsing config flags", err)
	}
}

func bindEnv() {
	viper.SetEnvPrefix("SH_ID_PLATFORM")
	_ = viper.BindEnv("ServerUrl", "SH_ID_PLATFORM_SERVER_URL")
	_ = viper.BindEnv("ServerPort", "SH_ID_PLATFORM_SERVER_PORT")
	_ = viper.BindEnv("Database.URL", "SH_ID_PLATFORM_DATABASE_URL")
	viper.AutomaticEnv()
}

func getWorkingDirectory() string {
	dir, _ := os.Getwd()
	path := strings.Split(dir, "sh-id-platform")
	return path[0] + "sh-id-platform/"
}
