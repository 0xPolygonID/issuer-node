package api

import (
	"context"
	b64 "encoding/base64"
	"errors"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// CreateKey is the handler for the POST /keys endpoint.
func (s *Server) CreateKey(ctx context.Context, request CreateKeyRequestObject) (CreateKeyResponseObject, error) {
	if string(request.Body.KeyType) != string(BJJ) && string(request.Body.KeyType) != string(ETH) {
		log.Error(ctx, "invalid key type. BJJ and ETH Keys are supported")
		return CreateKey400JSONResponse{
			N400JSONResponse{
				Message: "invalid key type. BJJ and ETH Keys are supported",
			},
		}, nil
	}

	keyID, err := s.keyService.CreateKey(ctx, request.Identifier.did(), kms.KeyType(request.Body.KeyType))
	if err != nil {
		log.Error(ctx, "creating key", "err", err)
		return CreateKey500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}
	return CreateKey201JSONResponse{
		KeyID: keyID.ID,
	}, nil
}

// GetKey is the handler for the GET /keys/{keyID} endpoint.
func (s *Server) GetKey(ctx context.Context, request GetKeyRequestObject) (GetKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.KeyID)
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
		if errors.Is(err, ports.ErrInvalidKeyType) {
			return GetKey400JSONResponse{
				N400JSONResponse{
					Message: "invalid key type",
				},
			}, nil
		}

		if errors.Is(err, ports.ErrKeyNotFound) {
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
		KeyID:           encodedKeyID,
		KeyType:         KeyKeyType(key.KeyType),
		PublicKey:       key.PublicKey,
		IsAuthCoreClaim: key.HasAssociatedAuthCoreClaim,
	}, nil
}

// GetKeys is the handler for the GET /keys endpoint.
func (s *Server) GetKeys(ctx context.Context, request GetKeysRequestObject) (GetKeysResponseObject, error) {
	keys, err := s.keyService.GetAll(ctx, request.Identifier.did())
	if err != nil {
		log.Error(ctx, "getting keys", "err", err)
		return GetKeys500JSONResponse{
			N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	var keysResponse GetKeys200JSONResponse
	for _, key := range keys {
		encodedKeyID := b64.StdEncoding.EncodeToString([]byte(key.KeyID))
		keysResponse = append(keysResponse, Key{
			KeyID:           encodedKeyID,
			KeyType:         KeyKeyType(key.KeyType),
			PublicKey:       key.PublicKey,
			IsAuthCoreClaim: key.HasAssociatedAuthCoreClaim,
		})
	}
	return keysResponse, nil
}

// DeleteKey is the handler for the DELETE /keys/{keyID} endpoint.
func (s *Server) DeleteKey(ctx context.Context, request DeleteKeyRequestObject) (DeleteKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.KeyID)
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
		if errors.Is(err, ports.ErrAuthCredentialNotRevoked) {
			log.Error(ctx, "delete key. Auth core claim not revoked", "err", err)
			return DeleteKey400JSONResponse{
				N400JSONResponse{
					Message: "associated auth credential is not revoked",
				},
			}, nil
		}

		if errors.Is(err, ports.ErrKeyNotFound) {
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
