package services

import (
	"context"
	"fmt"
	"github.com/polygonid/sh-id-platform/internal/jsonschema"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"strconv"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

type CredentialLinkAttributeError struct {
	message string
}

func newCredentialLinkAttributeError(message string) error {
	return &CredentialLinkAttributeError{
		message: message,
	}
}

func (e *CredentialLinkAttributeError) Error() string {
	return e.message
}

// Link - represents a link in the issuer node
type Link struct {
	linkRepository   ports.LinkRepository
	schemaRepository ports.SchemaRepository
	loaderFactory    loader.Factory
}

// NewLinkService - constructor
func NewLinkService(linkRepository ports.LinkRepository) ports.LinkService {
	return &Link{
		linkRepository: linkRepository,
		loaderFactory:  loader.HTTPFactory,
	}
}

// Save - save a new credential
func (ls *Link) Save(ctx context.Context, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID, credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error) {
	link := domain.NewLink(did, maxIssuance, validUntil, schemaID, credentialExpiration, credentialSignatureProof, credentialMTPProof, credentialAttributes)

	schema, err := ls.schemaRepository.GetByID(ctx, schemaID)
	if err != nil {
		return nil, err
	}

	remoteSchema, err := jsonschema.Load(ctx, ls.loaderFactory(schema.URL))
	if err != nil {
		log.Error(ctx, "loading jsonschema", "err", err, "jsonschema", schema.URL)
		return nil, ErrLoadingSchema
	}
	schemaAttributes, err := remoteSchema.AttributeNames()
	if err != nil {
		log.Error(ctx, "processing jsonschema", "err", err, "jsonschema", schema.URL)
		return nil, ErrProcessSchema
	}

	for _, attributeLink := range credentialAttributes {
		attributeLinkName := attributeLink.Name
		attributeLinkValue := attributeLink.Value
		index := findIndexForSchemaAttribute(schemaAttributes, attributeLinkName)
		isValid := validateCredentialLinkAttribute(schemaAttributes[index], attributeLinkValue)
		if !isValid {
			return nil, newCredentialLinkAttributeError(fmt.Sprintf("the attribute %s is not valid", attributeLinkName))
		}
		schemaAttributes = removeIndex(schemaAttributes, index)
	}

	if len(schemaAttributes) != 1 {
		return nil, newCredentialLinkAttributeError("the number of attributes is not valid")
	}

	_, err = ls.linkRepository.Save(ctx, link)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func findIndexForSchemaAttribute(attributes jsonschema.Attributes, name string) int {
	for i, attribute := range attributes {
		if attribute.ID == name {
			return i
		}
	}
	return -1
}

func validateCredentialLinkAttribute(schemaAttribute jsonschema.Attribute, attributeLinkValue string) bool {
	if schemaAttribute.Type == "string" {
		return true
	}

	if schemaAttribute.Type == "integer" {
		_, err := strconv.Atoi(attributeLinkValue)
		if err != nil {
			return false
		}
		return true
	}

	if schemaAttribute.Type == "boolean" {
		_, err := strconv.ParseBool(attributeLinkValue)
		if err != nil {
			return false
		}
		return true
	}

	return false
}

func removeIndex(s []jsonschema.Attribute, index int) []jsonschema.Attribute {
	return append(s[:index], s[index+1:]...)
}

// Delete - delete a link by id
func (ls *Link) Delete(ctx context.Context, id uuid.UUID) error {
	return ls.linkRepository.Delete(ctx, id)
}
