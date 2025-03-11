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

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/iden3comm/v2/protocol"
	"gopkg.in/yaml.v3"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// OptionConfigIDType is a custom type to represent the payment option id
type OptionConfigIDType int

// Config is a map of payment option id and chain config
type Config map[OptionConfigIDType]ChainConfig

// ChainConfig is the configuration for a chain
type ChainConfig struct {
	ChainID       int
	PaymentRails  common.Address
	PaymentOption PaymentOptionConfig
}

// PaymentOptionConfig is the configuration for a payment option
type PaymentOptionConfig struct {
	Name            string
	Type            protocol.PaymentRequestType
	ContractAddress common.Address
	Features        []protocol.PaymentFeatures `json:"features,omitempty"`
	Decimals        int
}

// configDTO is the data transfer object for the configuration. It maps to payment configuration file
type configDTO map[int]chainConfigDTO

type chainConfigDTO struct {
	PaymentRails   string                   `yaml:"PaymentRails" json:"paymentRails"`
	PaymentOptions []paymentOptionConfigDTO `yaml:"PaymentOptions" json:"paymentOptions"`
}

type paymentOptionConfigDTO struct {
	ID              int                         `yaml:"ID" json:"id"`
	Name            string                      `yaml:"Name" json:"name"`
	Type            protocol.PaymentRequestType `yaml:"Type" json:"type"`
	ContractAddress string                      `yaml:"ContractAddress,omitempty" json:"contractAddress,omitempty"`
	Features        []string                    `yaml:"Features,omitempty" json:"features,omitempty"`
	Decimals        int                         `yaml:"Decimals" json:"decimals"`
}

// CustomDecoder wraps yaml.Decoder to add custom functionality
type configDecoder struct {
	decoder *yaml.Decoder
}

// NewCustomDecoder creates a new CustomDecoder
func newConfigDecoder(r io.Reader) *configDecoder {
	return &configDecoder{
		decoder: yaml.NewDecoder(r),
	}
}

// Decode parse the payment settings yaml file, do some checks and returns
// a Config object with curated data
// Config object is created from the yaml using the configDTO struct.
// Information is the same but formatted in a more usable way
func (d *configDecoder) Decode() (*Config, error) {
	var dto configDTO
	var cfg Config

	err := d.decoder.Decode(&dto)
	if err != nil {
		return nil, fmt.Errorf("cannot decode payment settings file: %w", err)
	}
	cfg = make(Config)
	// Converting the dto to a Config object
	for id, chainConfig := range dto {
		if !common.IsHexAddress(chainConfig.PaymentRails) {
			return nil, fmt.Errorf("invalid payment rails address: %s", chainConfig.PaymentRails)
		}
		for _, option := range chainConfig.PaymentOptions {
			if _, found := cfg[OptionConfigIDType(option.ID)]; found {
				return nil, fmt.Errorf("duplicate payment option id: %d", id)
			}
			if !common.IsHexAddress(option.ContractAddress) && strings.TrimSpace(option.ContractAddress) != "" {
				return nil, fmt.Errorf("invalid PaymentRails address: %s", chainConfig.PaymentRails)
			}
			var features []protocol.PaymentFeatures
			if len(option.Features) > 0 {
				features = make([]protocol.PaymentFeatures, len(option.Features))
				for i, feature := range option.Features {
					features[i] = protocol.PaymentFeatures(feature)
				}
			}
			cfg[OptionConfigIDType(option.ID)] = ChainConfig{
				ChainID:      id,
				PaymentRails: common.HexToAddress(chainConfig.PaymentRails),
				PaymentOption: PaymentOptionConfig{
					Name:            option.Name,
					Type:            option.Type,
					ContractAddress: common.HexToAddress(option.ContractAddress),
					Features:        features,
					Decimals:        option.Decimals,
				},
			}
		}
	}
	return &cfg, nil
}

// SettingsFromConfig returns the settings from the configuration
// It reads the settings from the file if the path is provided or from the base64 encoded file injected
// into the configuration via an environment variable
func SettingsFromConfig(ctx context.Context, cfg *config.Payments) (*Config, error) {
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
func SettingsFromReader(reader io.Reader) (*Config, error) {
	return newConfigDecoder(reader).Decode()
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
