package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
)

type discovery struct {
	mediatypeManager ports.MediatypeManager
}

func NewDiscovery(mediatypeManager ports.MediatypeManager) *discovery {
	d := &discovery{
		mediatypeManager: mediatypeManager,
	}
	return d
}

func (c *discovery) Agent(ctx context.Context, req *ports.AgentRequest) (*domain.Agent, error) {
	fmt.Println("discovery.Agent")
	if !c.mediatypeManager.AllowMediaType(req.Type, req.Typ) {
		err := fmt.Errorf("unsupported media type '%s' for message type '%s'", req.Typ, req.Type)
		log.Error(ctx, "agent: unsupported media type", "err", err)
		return nil, err
	}

	queries := &protocol.DiscoverFeatureQueriesMessageBody{}
	err := json.Unmarshal(req.Body, queries)
	if err != nil {
		log.Error(ctx, "unmarshalling agent body", "err", err)
		return nil, fmt.Errorf("invalid discover feature queries request body: %w", err)
	}

	disclosures := []protocol.DiscoverFeatureDisclosure{}

	for _, query := range queries.Queries {
		switch query.FeatureType {
		case protocol.DiscoveryProtocolFeatureTypeAccept:
			acceptDisclosures, err := c.handleAccept(ctx)
			if err != nil {
				return nil, err
			}
			disclosures = append(disclosures, acceptDisclosures...)
			break
		case protocol.DiscoveryProtocolFeatureTypeGoalCode:
			goalCodeDisclosures := c.handleGoalCode(ctx)
			disclosures = append(disclosures, goalCodeDisclosures...)
			break
		case protocol.DiscoveryProtocolFeatureTypeProtocol:
			protocolDisclosures := c.handleProtocol(ctx)
			disclosures = append(disclosures, protocolDisclosures...)
			break
		case protocol.DiscoveryProtocolFeatureTypeHeader:
			headertDisclosures := c.handleHeader(ctx)
			disclosures = append(disclosures, headertDisclosures...)
			break
		}

	}

	var from, to string
	if req.IssuerDID != nil {
		from = req.IssuerDID.String()
	}
	if req.UserDID != nil {
		to = req.UserDID.String()
	}

	return &domain.Agent{
		ID:       uuid.NewString(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.DiscoverFeatureDiscloseMessageType,
		ThreadID: req.ThreadID,
		Body: protocol.DiscoverFeatureDiscloseMessageBody{
			Disclosures: disclosures,
		},
		From: from,
		To:   to,
	}, nil
}

func (d *discovery) handleAccept(ctx context.Context) ([]protocol.DiscoverFeatureDisclosure, error) {
	disclosures := []protocol.DiscoverFeatureDisclosure{}
	return disclosures, nil
}

func (d *discovery) handleProtocol(ctx context.Context) []protocol.DiscoverFeatureDisclosure {
	disclosures := []protocol.DiscoverFeatureDisclosure{}
	return disclosures
}

func (d *discovery) handleGoalCode(ctx context.Context) []protocol.DiscoverFeatureDisclosure {
	disclosures := []protocol.DiscoverFeatureDisclosure{}
	return disclosures
}

func (d *discovery) handleHeader(ctx context.Context) []protocol.DiscoverFeatureDisclosure {
	headers := []string{
		"id",
		"typ",
		"type",
		"thid",
		"body",
		"from",
		"to",
		"created_time",
		"expires_time",
	}

	disclosures := []protocol.DiscoverFeatureDisclosure{}

	for _, header := range headers {
		disclosures = append(disclosures, protocol.DiscoverFeatureDisclosure{
			FeatureType: protocol.DiscoveryProtocolFeatureTypeHeader,
			ID:          header,
		})
	}

	return disclosures
}

func wildcardToRegExp(match string) *regexp.Regexp {
	// Escape special regex characters and replace `*` with `.*`
	regexPattern := regexp.QuoteMeta(match)
	regexPattern = strings.ReplaceAll(regexPattern, "\\*", ".*")
	regExp, _ := regexp.Compile("^" + regexPattern + "$")
	return regExp
}

func (d *discovery) handleMatch(ctx context.Context, disclosures []protocol.DiscoverFeatureDisclosure, match string) []protocol.DiscoverFeatureDisclosure {
	if match == "" || match == "*" {
		return disclosures
	}

	regExp := wildcardToRegExp(match)
	var filtered []protocol.DiscoverFeatureDisclosure
	for _, disclosure := range disclosures {
		if regExp.MatchString(disclosure.ID) {
			filtered = append(filtered, disclosure)
		}
	}
	return filtered
}
