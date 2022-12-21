package api

import (
	"context"
	"math/rand"
)

type Server struct{}

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
