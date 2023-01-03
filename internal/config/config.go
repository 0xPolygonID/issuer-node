package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Configuration struct {
	ServerPort int
	Database   Database `mapstructure:"Database"`
	KeyStore   KeyStore `mapstructure:"KeyStore"`
}

type Database struct {
	Url string
}

type KeyStore struct {
	Address              string
	Token                string
	PluginIden3MountPath string
}

// Load -
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
		viper.SetConfigType("toml")
		if fileName == "" {
			viper.SetConfigName("config")
		} else {
			viper.SetConfigName(fileName)
		}
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
		fmt.Printf("Error parsing flags: %s", err.Error())
	}
}

func bindEnv() {
	viper.SetEnvPrefix("SH_ID_PLATFORM")
	_ = viper.BindEnv("ServerPort", "SH_ID_PLATFORM_SERVER_PORT")
	_ = viper.BindEnv("Database.Url", "SH_ID_PLATFORM_DATABASE_URL")
	viper.AutomaticEnv()
}

func getWorkingDirectory() string {
	dir, _ := os.Getwd()
	path := strings.Split(dir, "sh-id-platform")
	return path[0] + "sh-id-platform/"
}
