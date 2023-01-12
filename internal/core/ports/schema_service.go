package ports

import (
	"context"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"
)

// SchemaService is the interface implemented by the Schema service
type SchemaService interface {
	LoadSchema(ctx context.Context, url string) (jsonSuite.Schema, error)
	Process(ctx context.Context, schemaURL, credentialType string, credential verifiable.W3CCredential, opts *processor.CoreClaimOptions) (*core.Claim, error)
}
