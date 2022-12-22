package services

import (
	ports2 "github.com/polygonid/sh-id-platform/internal/core/ports"
)

type identity struct {
	identityRepository ports2.IndentityRepository
}

func NewIdentity(repository ports2.IndentityRepository) ports2.IndentityService {
	return &identity{
		identityRepository: repository,
	}
}

func (i *identity) Create() error {
	return nil
}
