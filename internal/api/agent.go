package api

import (
	"context"

	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// Agent is the controller to fetch credentials from mobile
func (s *Server) Agent(ctx context.Context, request AgentRequestObject) (AgentResponseObject, error) {
	if request.Body == nil || *request.Body == "" {
		log.Debug(ctx, "agent empty request")
		return Agent400JSONResponse{N400JSONResponse{"cannot proceed with an empty request"}}, nil
	}

	basicMessage, mediatype, err := s.packageManager.Unpack([]byte(*request.Body))
	if err != nil {
		log.Debug(ctx, "agent bad request", "err", err, "body", *request.Body)
		return Agent400JSONResponse{N400JSONResponse{"cannot proceed with the given request"}}, nil
	}

	var agent *domain.Agent

	if basicMessage.Type == protocol.DiscoverFeatureQueriesMessageType {
		req, err := ports.NewDiscoveryAgentRequest(basicMessage)
		if err != nil {
			log.Error(ctx, "agent parsing request", "err", err)
			return Agent400JSONResponse{N400JSONResponse{err.Error()}}, nil
		}

		agent, err = s.discoveryService.Agent(ctx, req)
		if err != nil {
			log.Error(ctx, "agent error", "err", err)
			return Agent400JSONResponse{N400JSONResponse{err.Error()}}, nil
		}

	} else {

		req, err := ports.NewAgentRequest(basicMessage)
		if err != nil {
			log.Error(ctx, "agent parsing request", "err", err)
			return Agent400JSONResponse{N400JSONResponse{err.Error()}}, nil
		}

		agent, err = s.claimService.Agent(ctx, req, mediatype)
		if err != nil {
			log.Error(ctx, "agent error", "err", err)
			return Agent400JSONResponse{N400JSONResponse{err.Error()}}, nil
		}
	}

	return Agent200JSONResponse{
		Body:     agent.Body,
		From:     agent.From,
		Id:       agent.ID,
		ThreadID: agent.ThreadID,
		To:       agent.To,
		Typ:      string(agent.Typ),
		Type:     string(agent.Type),
	}, nil
}

// AgentV1 is the controller to fetch credentials from mobile
func (s *Server) AgentV1(ctx context.Context, request AgentV1RequestObject) (AgentV1ResponseObject, error) {
	if request.Body == nil || *request.Body == "" {
		log.Debug(ctx, "agent empty request")
		return AgentV1400JSONResponse{N400JSONResponse{"cannot proceed with an empty request"}}, nil
	}

	basicMessage, mediatype, err := s.packageManager.Unpack([]byte(*request.Body))
	if err != nil {
		log.Debug(ctx, "agent bad request", "err", err, "body", *request.Body)
		return AgentV1400JSONResponse{N400JSONResponse{"cannot proceed with the given request"}}, nil
	}

	req, err := ports.NewAgentRequest(basicMessage)
	if err != nil {
		log.Error(ctx, "agent parsing request", "err", err)
		return AgentV1400JSONResponse{N400JSONResponse{err.Error()}}, nil
	}

	agent, err := s.claimService.Agent(ctx, req, mediatype)
	if err != nil {
		log.Error(ctx, "agent error", "err", err)
		return AgentV1400JSONResponse{N400JSONResponse{err.Error()}}, nil
	}
	return AgentV1200JSONResponse{
		Body:     agent.Body,
		From:     agent.From,
		Id:       agent.ID,
		ThreadID: agent.ThreadID,
		To:       agent.To,
		Typ:      string(agent.Typ),
		Type:     string(agent.Type),
	}, nil
}
