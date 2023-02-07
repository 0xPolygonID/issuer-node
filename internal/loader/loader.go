package loader

import (
	"github.com/iden3/go-schema-processor/processor"
)

// Loader defines a Loader interface
type Loader interface {
	processor.SchemaLoader
}

// Factory defines the interface that a loader constructor should satisfy
type Factory func(url string) Loader
