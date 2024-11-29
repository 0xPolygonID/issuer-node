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
	if string(request.Body.KeyType) != string(BJJ) {
		log.Error(ctx, "create key. Invalid key type. BJJ and ETH Keys are supported")
		return CreateKey400JSONResponse{
			N400JSONResponse{
				Message: "Invalid key type. BJJ Keys are supported",
			},
		}, nil
	}

	keyID, err := s.keyService.CreateKey(ctx, request.Identifier.did(), kms.KeyType(request.Body.KeyType))
	if err != nil {
		log.Error(ctx, "add key. Creating key", "err", err)
		return CreateKey500JSONResponse{
			N500JSONResponse{
				Message: "internal error",
			},
		}, err
	}

	encodedKeyID := b64.StdEncoding.EncodeToString([]byte(keyID.ID))
	return CreateKey201JSONResponse{
		KeyID: encodedKeyID,
	}, nil
}

// GetKey is the handler for the GET /keys/{keyID} endpoint.
func (s *Server) GetKey(ctx context.Context, request GetKeyRequestObject) (GetKeyResponseObject, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(request.KeyID)
	if err != nil {
		log.Error(ctx, "get key. Decoding key id", "err", err)
		return GetKey400JSONResponse{
			N400JSONResponse{
				Message: "invalid key id",
			},
		}, nil
	}

	key, err := s.keyService.Get(ctx, request.Identifier.did(), string(decodedKeyID))
	if err != nil {
		log.Error(ctx, "get key. Getting key", "err", err)
		if errors.Is(err, ports.ErrInvalidKeyType) {
			return GetKey400JSONResponse{
				N400JSONResponse{
					Message: "invalid key type",
				},
			}, nil
		}

		return GetKey500JSONResponse{
			N500JSONResponse{
				Message: "internal error",
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
		log.Error(ctx, "get keys. Getting keys", "err", err)
		return GetKeys500JSONResponse{
			N500JSONResponse{
				Message: "internal error",
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
