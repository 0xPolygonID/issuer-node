package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
)

type discovery struct {
	mediatypeManager ports.MediatypeManager
	packerManager    *iden3comm.PackageManager
}

// NewDiscovery is a constructor for the discovery service
func NewDiscovery(mediatypeManager ports.MediatypeManager, packerManager *iden3comm.PackageManager) *discovery {
	d := &discovery{
		mediatypeManager: mediatypeManager,
		packerManager:    packerManager,
	}
	return d
}

func (c *discovery) Agent(ctx context.Context, req *ports.AgentRequest) (*domain.Agent, error) {
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
		var disclosuresToAppend []protocol.DiscoverFeatureDisclosure
		switch query.FeatureType {
		case protocol.DiscoveryProtocolFeatureTypeAccept:
			disclosuresToAppend, err = c.handleAccept(ctx)
			if err != nil {
				return nil, err
			}
		case protocol.DiscoveryProtocolFeatureTypeGoalCode:
			disclosuresToAppend = c.handleGoalCode(ctx)
		case protocol.DiscoveryProtocolFeatureTypeProtocol:
			disclosuresToAppend = c.handleProtocol(ctx)
		case protocol.DiscoveryProtocolFeatureTypeHeader:
			disclosuresToAppend = c.handleHeader(ctx)
		}
		disclosuresToAppend = c.handleMatch(ctx, disclosuresToAppend, query.Match)
		disclosures = append(disclosures, disclosuresToAppend...)
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

func (d *discovery) handleAccept(_ context.Context) ([]protocol.DiscoverFeatureDisclosure, error) {
	disclosures := []protocol.DiscoverFeatureDisclosure{}

	profiles := d.packerManager.GetSupportedProfiles()
	for _, profile := range profiles {
		disclosures = append(disclosures, protocol.DiscoverFeatureDisclosure{
			FeatureType: protocol.DiscoveryProtocolFeatureTypeAccept,
			ID:          profile,
		})
	}
	return disclosures, nil
}

func (d *discovery) handleProtocol(_ context.Context) []protocol.DiscoverFeatureDisclosure {
	disclosures := []protocol.DiscoverFeatureDisclosure{}
	return disclosures
}

func (d *discovery) handleGoalCode(_ context.Context) []protocol.DiscoverFeatureDisclosure {
	disclosures := []protocol.DiscoverFeatureDisclosure{}
	return disclosures
}

func (d *discovery) handleHeader(_ context.Context) []protocol.DiscoverFeatureDisclosure {
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

func (d *discovery) handleMatch(_ context.Context, disclosures []protocol.DiscoverFeatureDisclosure, match string) []protocol.DiscoverFeatureDisclosure {
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
