package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-schema-processor/verifiable"
)

// RevocationService is the interface implemented by the RevocationService service
type RevocationService interface {
	Status(ctx context.Context, credStatus interface{}, issuerDID *core.DID) (*verifiable.RevocationStatus, error)
}
