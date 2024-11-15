package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
)

// DefaultDisplayMethodErr is the error returned when trying to delete the default display method
var DefaultDisplayMethodErr = errors.New("cannot delete default display method")

// DisplayMethod represents the display method service
type DisplayMethod struct {
	displayMethodRepository ports.DisplayMethodRepository
}

// NewDisplayMethod creates a new display method service
func NewDisplayMethod(displayMethodRepository ports.DisplayMethodRepository) *DisplayMethod {
	return &DisplayMethod{displayMethodRepository: displayMethodRepository}
}

// Save stores the display method
func (dm *DisplayMethod) Save(ctx context.Context, identityDID w3c.DID, name, url string, isDefault bool) (*uuid.UUID, error) {
	displayMethod := domain.NewDisplayMethod(uuid.New(), identityDID, name, url, isDefault)
	return dm.displayMethodRepository.Save(ctx, displayMethod)
}

// Update updates the display method with the given id
func (dm *DisplayMethod) Update(ctx context.Context, identityDID w3c.DID, id uuid.UUID, name, url *string, isDefault *bool) (*uuid.UUID, error) {
	displayMethodToUpdate, err := dm.GetByID(ctx, identityDID, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		displayMethodToUpdate.Name = *name
	}

	if url != nil {
		displayMethodToUpdate.URL = *url
	}

	if isDefault != nil {
		displayMethodToUpdate.IsDefault = *isDefault
	}

	return dm.displayMethodRepository.Save(ctx, *displayMethodToUpdate)
}

// GetByID returns the display method with the given id
func (dm *DisplayMethod) GetByID(ctx context.Context, identityDID w3c.DID, id uuid.UUID) (*domain.DisplayMethod, error) {
	return dm.displayMethodRepository.GetByID(ctx, identityDID, id)
}

// GetAll returns all display methods for the given identity
func (dm *DisplayMethod) GetAll(ctx context.Context, identityDID w3c.DID, filter ports.DisplayMethodFilter) ([]domain.DisplayMethod, uint, error) {
	return dm.displayMethodRepository.GetAll(ctx, identityDID, filter)
}

// Delete removes the display method with the given id
func (dm *DisplayMethod) Delete(ctx context.Context, identityDID w3c.DID, id uuid.UUID) error {
	displayMethodToDelete, err := dm.GetByID(ctx, identityDID, id)
	if err != nil {
		return err
	}

	if displayMethodToDelete.IsDefault {
		return DefaultDisplayMethodErr
	}

	return dm.displayMethodRepository.Delete(ctx, identityDID, id)
}

// GetDefault returns the default display method for the given identity
func (dm *DisplayMethod) GetDefault(ctx context.Context, identityDID w3c.DID) (*domain.DisplayMethod, error) {
	return dm.displayMethodRepository.GetDefault(ctx, identityDID)
}
