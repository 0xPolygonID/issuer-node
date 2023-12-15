package tests

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/polygonid/issuer-node/internal/config"
	"github.com/polygonid/issuer-node/internal/db"
	"github.com/polygonid/issuer-node/internal/db/schema"
)

const (
	defaultTimeOut = 40
)

// NewTestStorage tests the storage
func NewTestStorage(cfg *config.Configuration) (*db.Storage, func(), error) {
	noopTeardown := func() {}
	if cfg.Database.URL == "" {
		return nil, noopTeardown, errors.New("testdb: no connection string")
	}

	tempDBName := "sh_id_platform_test_" + time.Now().UTC().Format("20060102150405.999999999")
	tempURL, err := url.Parse(cfg.Database.URL + "/" + tempDBName + "?sslmode=disable")
	if err != nil {
		return nil, noopTeardown, fmt.Errorf("connection string is invalid: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeOut*time.Second)
	defer cancel()

	storage, err := db.NewStorage(cfg.Database.URL)
	if err != nil {
		return nil, noopTeardown, fmt.Errorf("can't connect to database: %v", err)
	}

	_, err = storage.Pgx.Exec(ctx, fmt.Sprintf(`create database "%s";`, tempDBName))
	if err != nil {
		return nil, noopTeardown, fmt.Errorf("failed to create database (%s): %v", tempDBName, err)
	}

	if err := schema.Migrate(tempURL.String()); err != nil {
		return nil, noopTeardown, fmt.Errorf("can't migrate database %v", err)
	}

	teardown := func() {
		_ = storage.Close()
	}

	_ = storage.Close()

	storage, err = db.NewStorage(tempURL.String())
	if err != nil {
		return nil, noopTeardown, fmt.Errorf("can't connect to database: %v", err)
	}

	return storage, teardown, nil
}
