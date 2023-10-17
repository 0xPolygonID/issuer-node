package ports

import (
	"context"
	"encoding/json"

	"github.com/iden3/go-rapidsnark/types"
)

// ZKGenerator interface
type ZKGenerator interface {
	Generate(ctx context.Context, inputs json.RawMessage, circuitName string) (*types.ZKProof, error)
}
