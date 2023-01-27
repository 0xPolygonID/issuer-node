package ports

import (
	"context"
	"encoding/json"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// ZKGenerator interface
type ZKGenerator interface {
	Generate(ctx context.Context, inputs json.RawMessage, circuitName string) (*domain.FullProof, error)
}
