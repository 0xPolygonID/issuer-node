package loader

import "github.com/iden3/go-schema-processor/loaders"

// HTTPFactory returns an http loader
func HTTPFactory(u string) Loader {
	return &loaders.HTTP{URL: u}
}
