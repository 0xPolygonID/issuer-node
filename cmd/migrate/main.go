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
	cfg, err := config.Load("")
	if err != nil {
		log.Error(context.Background(), "cannot load config", err)
	}
	// Context with log
	ctx := log.NewContext(context.Background(), cfg.Log.Level, cfg.Log.Mode, os.Stdout)
	log.Debug(ctx, "database", "url", cfg.Database.URL)

	if err := schema.Migrate(cfg.Database.URL); err != nil {
		log.Error(ctx, "error migrating database", err)
		return
	}

	log.Info(ctx, "migration done!")
}
