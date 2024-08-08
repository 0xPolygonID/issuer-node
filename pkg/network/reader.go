package network

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// GetReaderFromConfig returns a reader for the network resolver settings file
func GetReaderFromConfig(cfg *config.Configuration, ctx context.Context) (io.Reader, error) {
	var reader io.Reader
	var err error
	if cfg.NetworkResolverPath != "" {
		reader, err = readFileFromPath(ctx, cfg.NetworkResolverPath)
	} else {
		reader, err = readFile(ctx, cfg.NetworkResolverFile)
	}
	return reader, err
}

// ReadFileFromPath is a function that returns a reader for the resolver settings file
func readFileFromPath(ctx context.Context, resolverSettingsPath string) (io.Reader, error) {
	if _, err := os.Stat(resolverSettingsPath); errors.Is(err, os.ErrNotExist) {
		log.Info(ctx, "resolver settings file not found", "path", resolverSettingsPath)
		log.Info(ctx, "issuer node wil not run supporting multi chain feature")
		return nil, fmt.Errorf("resolver settings file not found: %s", resolverSettingsPath)
	}

	if info, _ := os.Stat(resolverSettingsPath); info.Size() == 0 {
		log.Info(ctx, "resolver settings file is empty")
		return nil, fmt.Errorf("resolver settings file is empty: %s", resolverSettingsPath)
	}

	f, err := os.Open(filepath.Clean(resolverSettingsPath))
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ReadFile is a function that returns a reader for the encoded (base64) file
func readFile(ctx context.Context, encodedContentFile string) (io.Reader, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedContentFile)
	if err != nil {
		log.Error(ctx, "cannot decode base64 encoded file", "err", err)
		return nil, err
	}
	decodedString := string(decodedBytes)
	//nolint:all
	fmt.Println("file decoded: \n", decodedString)
	return strings.NewReader(decodedString), nil
}
