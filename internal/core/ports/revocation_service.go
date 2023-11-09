package ports

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
)

// RevocationService is the interface implemented by the RevocationService service
type RevocationService interface {
	Status(ctx context.Context, credStatus interface{}, issuerDID *w3c.DID, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error)
}
