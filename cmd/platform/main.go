package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load("internal/config")
	if err != nil {
		log.Error(context.Background(), "cannot load config", err)
		return
	}
	// Context with log
	ctx := log.NewContext(context.Background(), cfg.Runtime.LogLevel, cfg.Runtime.LogMode, os.Stdout)

	repo := repositories.NewIdentity(db.NewSqlx(cfg.Database.URL))
	service := services.NewIdentity(repo)

	mux := chi.NewRouter()
	api.HandlerFromMux(api.NewStrictHandler(api.NewServer(service), middlewares(ctx)), mux)
	api.RegisterStatic(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error(ctx, "Starting http server", err)
		}
	}()

	<-quit
	log.Info(ctx, "Shutting down")
}

func middlewares(ctx context.Context) []api.StrictMiddlewareFunc {
	return []api.StrictMiddlewareFunc{
		api.LogMiddleware(ctx),
	}
}
