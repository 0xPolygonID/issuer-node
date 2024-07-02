package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// ReadFile is a function that returns a reader for the resolver settings file
func ReadFile(ctx context.Context, resolverSettingsPath string) (io.Reader, error) {
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
