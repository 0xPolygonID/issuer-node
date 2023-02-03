package main

import (
	"context"
	"os"

	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/db/schema"
	"github.com/polygonid/sh-id-platform/internal/log"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	// Context with log
	ctx := log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout)
	log.Debug(ctx, "database", "url", cfg.Database.URL)

	if err := schema.Migrate(cfg.Database.URL); err != nil {
		log.Error(ctx, "error migrating database", err)
		return
	}

	log.Info(ctx, "migration done!")
}
