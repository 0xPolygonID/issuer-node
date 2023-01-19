package services_tests

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/db/tests"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/providers"
)

var (
	storage        *db.Storage
	vaultCli       *api.Client
	bjjKeyProvider kms.KeyProvider
	keyStore       *kms.KMS
)

//func TestMain(m *testing.M) {
//	os.Exit(testMain(m))
//}

func TestMain(m *testing.M) {
	ctx := context.Background()
	conn := lookupPostgresURL()
	if conn == "" {
		conn = "postgres://postgres:postgres@localhost:5435"
	}

	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", err)
		panic(err)
	}

	cfgForTesting := config.Configuration{
		Database: config.Database{
			URL: conn,
		},
		KeyStore: config.KeyStore{
			Address:              cfg.KeyStore.Address,
			Token:                cfg.KeyStore.Token,
			PluginIden3MountPath: cfg.KeyStore.PluginIden3MountPath,
		},
	}
	s, teardown, err := tests.NewTestStorage(&cfgForTesting)
	defer teardown()
	if err != nil {
		log.Error(ctx, "failed to acquire test database: %+v", err)
		// return 1
	}
	storage = s

	vaultCli, err = providers.NewVaultClient(cfgForTesting.KeyStore.Address, cfgForTesting.KeyStore.Token)
	if err != nil {
		log.Error(ctx, "failed to acquire vault client: %+v", err)
		// return 1
	}

	bjjKeyProvider, err = kms.NewVaultPluginIden3KeyProvider(vaultCli, cfgForTesting.KeyStore.PluginIden3MountPath, kms.KeyTypeBabyJubJub)
	if err != nil {
		log.Error(ctx, "failed to create Iden3 Key Provider: %+v", err)
		// return 1
	}

	keyStore = kms.NewKMS()
	err = keyStore.RegisterKeyProvider(kms.KeyTypeBabyJubJub, bjjKeyProvider)
	if err != nil {
		log.Error(ctx, "failed to register Key Provider: %+v", err)
		// return 1
	}

	// return m.Run()
	m.Run()
}

func lookupPostgresURL() string {
	con, ok := os.LookupEnv("POSTGRES_TEST_DATABASE")
	if !ok {
		return ""
	}
	return con
}
