package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"

	"github.com/polygonid/sh-id-platform/internal/db/schema"
	"github.com/polygonid/sh-id-platform/internal/log"

	_ "github.com/lib/pq"
)

// IssuerDatabaseUrl is the environment variable for the issuer database URL
const IssuerDatabaseUrl = "ISSUER_DATABASE_URL"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if os.Getenv(IssuerDatabaseUrl) == "" {
		err := godotenv.Load(".env-issuer")
		if err != nil {
			log.Error(ctx, "Error loading .env-issuer file")
		}
	}

	databaseUrl := os.Getenv(IssuerDatabaseUrl)
	log.Config(log.LevelDebug, log.LevelDebug, os.Stdout)
	log.Debug(ctx, "database", "url", databaseUrl)

	if err := schema.Migrate(databaseUrl); err != nil {
		log.Error(ctx, "error migrating database", "err", err)
		return
	}

	log.Info(ctx, "migration done!")
}
