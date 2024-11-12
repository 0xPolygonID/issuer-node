package payments

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// Settings holds the payments settings for the different chains
type Settings map[int]ChainSettings

// ChainSettings holds the settings for a chain
type ChainSettings struct {
	MCPayment string `yaml:"MCPayment"`
	ERC20     *ERC20 `yaml:"ERC20,omitempty"`
}

// ERC20 holds the settings for the ERC20 tokens
type ERC20 struct {
	USDT Token `yaml:"USDT"`
	USDC Token `yaml:"USDC"`
}

// Token holds the settings for a token
type Token struct {
	ContractAddress string   `yaml:"ContractAddress"`
	Features        []string `yaml:"Features"`
}

// SettingsFromConfig returns the settings from the configuration
// It reads the settings from the file if the path is provided or from the base64 encoded file injected
// into the configuration via an environment variable
func SettingsFromConfig(ctx context.Context, cfg *config.Payments) (*Settings, error) {
	var reader io.Reader
	var err error
	if cfg.SettingsPath != "" {
		reader, err = readFileFromPath(ctx, cfg.SettingsPath)
		if err != nil {
			log.Error(ctx, "cannot read settings file", "err", err)
			return nil, err
		}
		return SettingsFromReader(reader)
	}
	if settingsFileHasContent(cfg) {
		reader, err = readBase64FileContent(ctx, *cfg.SettingsFile)
		if err != nil {
			log.Error(ctx, "cannot read settings file", "err", err)
			return nil, err
		}
		return SettingsFromReader(reader)
	}
	return nil, errors.New("no payment settings file or payment file path provided")
}

func settingsFileHasContent(cfg *config.Payments) bool {
	return cfg.SettingsFile != nil && *cfg.SettingsFile != ""
}

// SettingsFromReader reads the settings from a reader
func SettingsFromReader(reader io.Reader) (*Settings, error) {
	var settings Settings
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// ReadFileFromPath is a function that returns a reader for the resolver settings file
func readFileFromPath(ctx context.Context, path string) (io.Reader, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Info(ctx, "payment settings file not found", "path", path)
		return nil, fmt.Errorf("payment settings file not found: %w", err)
	}

	if info, _ := os.Stat(path); info.Size() == 0 {
		log.Info(ctx, "payment settings file is empty")
		return nil, fmt.Errorf("payment settings file is empty: %s", path)
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// readBase64FileContent is a function that returns a reader for the encoded (base64) file
func readBase64FileContent(ctx context.Context, payload string) (io.Reader, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		log.Error(ctx, "cannot decode base64 encoded file", "err", err)
		return nil, err
	}
	return strings.NewReader(string(decodedBytes)), nil
}
