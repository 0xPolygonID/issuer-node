package api

import (
	"context"
	b64 "encoding/base64"
	"errors"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// CreateKey is the handler for the POST /keys endpoint.
func (s *Server) CreateKey(ctx context.Context, request CreateKeyRequestObject) (CreateKeyResponseObject, error) {
	if string(request.Body.KeyType) != string(KeyKeyTypeBabyjubJub) && string(request.Body.KeyType) != string(KeyKeyTypeSecp256k1) {
		log.Error(ctx, "invalid key type. babyjujJub and secp256k1 keys are supported")
		return CreateKey400JSONResponse{
			N400JSONResponse{
				Message: "invalid key type. babyjujJub and secp256k1 keys are supported are supported",
			},
		}, nil
	}

	if request.Body.Name == "" {
		log.Error(ctx, "name is required")
		return CreateKey400JSONResponse{
			N400JSONResponse{
				Message: "name is required",
			},
		}, nil
	}

	keyID, err := s.keyService.Create(ctx, request.Identifier.did(), convertKeyTypeFromRequest(request.Body.KeyType), request.Body.Name)
	if err != nil {
		log.Error(ctx, "creating key", "err", err)
		if errors.Is(err, repositories.ErrDuplicateKeyName) {
			log.Error(ctx, "duplicate key name", "err", err)
			return CreateKey400JSONResponse{
				N400JSONResponse{
					Message: "duplicate key name",
				},
			}, nil
		}
		return CreateKey500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}
	return CreateKey201JSONResponse{
		Id: keyID.ID,
	}, nil
}

// UpdateKey is the handler for the PATCH /keys/{keyID} endpoint.
func (s *Server) UpdateKey(ctx context.Context, request UpdateKeyRequestObject) (UpdateKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.Id)
	if err != nil {
		log.Error(ctx, "the key id can not be decoded from base64", "err", err)
		return UpdateKey400JSONResponse{
			N400JSONResponse{
				Message: "the key id can not be decoded from base64",
			},
		}, nil
	}

	if request.Body.Name == "" {
		log.Error(ctx, "name is required")
		return UpdateKey400JSONResponse{
			N400JSONResponse{
				Message: "name is required",
			},
		}, nil
	}

	err = s.keyService.Update(ctx, request.Identifier.did(), string(decodedKeyID), request.Body.Name)
	if err != nil {
		log.Error(ctx, "updating key", "err", err)
		if errors.Is(err, services.ErrKeyNotFound) {
			log.Error(ctx, "key not found", "err", err)
			return UpdateKey404JSONResponse{
				N404JSONResponse{
					Message: "key not found",
				},
			}, nil
		}
		if errors.Is(err, repositories.ErrDuplicateKeyName) {
			log.Error(ctx, "duplicate key name", "err", err)
			return UpdateKey400JSONResponse{
				N400JSONResponse{
					Message: "duplicate key name",
				},
			}, nil
		}
		return UpdateKey500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}
	return UpdateKey200JSONResponse{
		Message: "key updated",
	}, nil
}

// GetKey is the handler for the GET /keys/{keyID} endpoint.
func (s *Server) GetKey(ctx context.Context, request GetKeyRequestObject) (GetKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.Id)
	if err != nil {
		log.Error(ctx, "the key id can not be decoded from base64", "err", err)
		return GetKey400JSONResponse{
			N400JSONResponse{
				Message: "the key id can not be decoded from base64",
			},
		}, nil
	}

	key, err := s.keyService.Get(ctx, request.Identifier.did(), string(decodedKeyID))
	if err != nil {
		log.Error(ctx, "error getting the key", "err", err)
		if errors.Is(err, services.ErrInvalidKeyType) {
			return GetKey400JSONResponse{
				N400JSONResponse{
					Message: "invalid key type",
				},
			}, nil
		}

		if errors.Is(err, services.ErrKeyNotFound) {
			return GetKey404JSONResponse{
				N404JSONResponse{
					Message: "key not found",
				},
			}, nil
		}

		return GetKey500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}
	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(key.KeyID))
	return GetKey200JSONResponse{
		Id:               encodedKeyID,
		KeyType:          convertKeyTypeToResponse(key.KeyType),
		PublicKey:        key.PublicKey,
		IsAuthCredential: key.HasAssociatedAuthCredential,
		Name:             key.Name,
	}, nil
}

// GetKeys is the handler for the GET /keys endpoint.
func (s *Server) GetKeys(ctx context.Context, request GetKeysRequestObject) (GetKeysResponseObject, error) {
	const (
		defaultMaxResults = 50
		defaultPage       = 1
		minimumMaxResults = 10
	)
	filter := ports.KeyFilter{
		MaxResults: defaultMaxResults,
		Page:       defaultPage,
	}

	if request.Params.MaxResults != nil {
		if *request.Params.MaxResults < minimumMaxResults {
			filter.MaxResults = minimumMaxResults
		} else {
			filter.MaxResults = *request.Params.MaxResults
		}
	}

	if request.Params.Page != nil {
		filter.Page = *request.Params.Page
	}

	keys, total, err := s.keyService.GetAll(ctx, request.Identifier.did(), filter)
	if err != nil {
		log.Error(ctx, "getting keys", "err", err)
		return GetKeys500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	items := make([]Key, 0)
	for _, key := range keys {
		encodedKeyID := b64.StdEncoding.EncodeToString([]byte(key.KeyID))
		items = append(items, Key{
			Id:               encodedKeyID,
			KeyType:          convertKeyTypeToResponse(key.KeyType),
			PublicKey:        key.PublicKey,
			IsAuthCredential: key.HasAssociatedAuthCredential,
			Name:             key.Name,
		})
	}
	return GetKeys200JSONResponse{
		Items: items,
		Meta: PaginatedMetadata{
			Page:       filter.Page,
			MaxResults: filter.MaxResults,
			Total:      total,
		},
	}, nil
}

// DeleteKey is the handler for the DELETE /keys/{keyID} endpoint.
func (s *Server) DeleteKey(ctx context.Context, request DeleteKeyRequestObject) (DeleteKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.Id)
	if err != nil {
		log.Error(ctx, "the key id can not be decoded from base64", "err", err)
		return DeleteKey400JSONResponse{
			N400JSONResponse{
				Message: "the key id can not be decoded from base64",
			},
		}, nil
	}

	err = s.keyService.Delete(ctx, request.Identifier.did(), string(decodedKeyID))
	if err != nil {
		if errors.Is(err, services.ErrAuthCredentialNotRevoked) {
			log.Error(ctx, "delete key. Auth credential not revoked", "err", err)
			return DeleteKey400JSONResponse{
				N400JSONResponse{
					Message: "associated auth credential is not revoked",
				},
			}, nil
		}

		if errors.Is(err, services.ErrKeyAssociatedWithIdentity) {
			log.Error(ctx, "delete key. Key associated with identity", "err", err)
			return DeleteKey400JSONResponse{
				N400JSONResponse{
					Message: "key is associated with an identity",
				},
			}, nil
		}

		if errors.Is(err, services.ErrKeyNotFound) {
			log.Error(ctx, "key not found", "err", err)
			return DeleteKey404JSONResponse{
				N404JSONResponse{
					Message: "key not found",
				},
			}, nil
		}

		log.Error(ctx, "deleting key", "err", err)
		return DeleteKey500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return DeleteKey200JSONResponse{
		Message: "key deleted",
	}, nil
}

func convertKeyTypeToResponse(keyType kms.KeyType) KeyKeyType {
	if keyType == "BJJ" {
		return KeyKeyTypeBabyjubJub
	}
	return KeyKeyTypeSecp256k1
}

func convertKeyTypeFromRequest(keyType CreateKeyRequestKeyType) kms.KeyType {
	if string(keyType) == string(KeyKeyTypeBabyjubJub) {
		return "BJJ"
	}
	return "ETH"
}
