package services

import "github.com/polygonid/sh-id-platform/internal/indentity/core/ports"

type identity struct {
	identityRepository ports.IndentityRepository
}

func NewIdentity(repository ports.IndentityRepository) ports.IndentityService {
	return &identity{
		identityRepository: repository,
	}
}

func (i *identity) Create() error {
	return nil
}
