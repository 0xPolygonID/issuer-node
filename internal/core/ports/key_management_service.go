package ports

import (
	"context"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/kms"
)

type KeyManagementService interface {
	CreateKey(ctx context.Context, did *w3c.DID, keyType kms.KeyType) (kms.KeyID, error)
}
