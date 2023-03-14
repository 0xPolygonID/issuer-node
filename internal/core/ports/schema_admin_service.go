package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// SchemaAdminService defines the methods that Schema manager will expose.
type SchemaAdminService interface {
	ImportSchema(ctx context.Context, issuerDID core.DID, url string, sType string) (*domain.Schema, error)
}
