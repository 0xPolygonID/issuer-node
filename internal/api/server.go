package api

import (
	"context"
	"math/rand"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/polygonid/sh-id-platform/internal/log"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

type Server struct {
	indentityService ports.IndentityService
}

func NewServer(indentityService ports.IndentityService) *Server {
	return &Server{
		indentityService: indentityService,
	}
}

func (s *Server) Health(_ context.Context, _ HealthRequestObject) (HealthResponseObject, error) {
	return Health200JSONResponse{
		Cache: true,
		Db:    false,
	}, nil
}

func (s *Server) Ping(ctx context.Context, _ PingRequestObject) (PingResponseObject, error) {
	log.Info(ctx, "ping")
	return Ping201JSONResponse{Response: ToPointer("pong")}, nil
}

func (s *Server) Random(_ context.Context, _ RandomRequestObject) (RandomResponseObject, error) {
	randomMessages := []string{"might", "rays", "bicycle", "use", "certainly", "chicken", "tie", "rain", "tent", "straight", "excellent"}
	i := rand.Intn(len(randomMessages))
	randomResponses := []RandomResponseObject{
		Random400JSONResponse{N400JSONResponse{Message: &randomMessages[i]}},
		Random401JSONResponse{N401JSONResponse{Message: &randomMessages[i]}},
		Random402JSONResponse{N402JSONResponse{Message: &randomMessages[i]}},
		Random407JSONResponse{N407JSONResponse{Message: &randomMessages[i]}},
		Random500JSONResponse{N500JSONResponse{Message: &randomMessages[i]}},
	}
	return randomResponses[rand.Intn(len(randomResponses))], nil
}

func RegisterStatic(mux *chi.Mux) {
	mux.Get("/", documentation)
	mux.Get("/static/docs/api/api.yaml", swagger)
}

func documentation(w http.ResponseWriter, _ *http.Request) {
	writeFile("api/spec.html", w)
}

func swagger(w http.ResponseWriter, _ *http.Request) {
	writeFile("api/api.yaml", w)
}

func writeFile(path string, w http.ResponseWriter) {
	f, err := os.ReadFile(path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(f)
}
