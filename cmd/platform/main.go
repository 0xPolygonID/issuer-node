package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/polygonid/polygonid-api/sh-id-platform/internal/api"
)

func main() {
	mux := echo.New()
	api.RegisterHandlers(mux, api.NewStrictHandler(&api.Server{}, nil))

	addr := ":3000"
	server := &http.Server{
		Addr:    addr,
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
