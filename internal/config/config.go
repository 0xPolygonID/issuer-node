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

// Configuration holds the project configuration
type Configuration struct {
	ServerPort int
	Database   Database `mapstructure:"Database"`
}

// Database has the database configuration
type Database struct {
	URL string
}

// Load loads the configuraion from a file
func Load(path string) (*Configuration, error) {
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
		viper.AddConfigPath(path)
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	config := new(Configuration)

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
	_ = viper.BindEnv("ServerPort", "SH_ID_PLATFORM_SERVER_PORT")
	_ = viper.BindEnv("Database.URL", "SH_ID_PLATFORM_DATABASE_URL")
	viper.AutomaticEnv()
}
