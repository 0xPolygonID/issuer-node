package api

import (
	"context"
	"math/rand"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

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

func (s *Server) Ping(_ context.Context, _ PingRequestObject) (PingResponseObject, error) {
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

func documentation(ctx echo.Context) error {
	f, err := os.ReadFile("api/spec.html")
	if err != nil {
		return ctx.String(http.StatusNotFound, "not found")
	}
	return ctx.HTMLBlob(http.StatusOK, f)
}

func RegisterStatic(e *echo.Echo) {
	e.GET("/", documentation)
	e.GET("/static/docs/api/api.yaml", func(ctx echo.Context) error {
		f, err := os.ReadFile("api/api.yaml")
		if err != nil {
			return ctx.String(http.StatusNotFound, "not found")
		}

		return ctx.HTMLBlob(http.StatusOK, f)
	})
}

func (s *Server) CreateIdentity(ctx context.Context, request CreateIdentityRequestObject) (CreateIdentityResponseObject, error) {
	identity, err := s.indentityService.Create(ctx, "http://localhost:3001")
	if err != nil {
		return nil, err
	}
	return CreateIdentity201JSONResponse{
		Identifier: &identity.Identifier,
		Immutable:  &identity.Immutable,
		Relay:      &identity.Relay,
	}, nil
}
