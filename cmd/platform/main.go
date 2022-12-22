package main

import (
	"fmt"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/polygonid/sh-id-platform/internal/api"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/db"
)

func main() {
	cfg, err := config.Load("internal/config")
	if err != nil {
		log.Fatal(err)
	}

	repo := repositories.NewIdentity(db.NewSqlx(cfg.Database.Url))
	service := services.NewIdentity(repo)

	mux := echo.New()
	api.RegisterStatic(mux)
	api.RegisterHandlers(mux, api.NewStrictHandler(api.NewServer(service), nil))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: mux,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting http server: %w", err)
		}
	}()

	<-quit
	log.Info("Shutting down")
}
