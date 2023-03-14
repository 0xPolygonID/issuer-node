package loader

import (
	"context"
	"os"
)

type file struct {
	path string
}

// Load returns a json schema from the local file system
func (l *file) Load(_ context.Context) (schema []byte, extension string, err error) {
	dat, err := os.ReadFile(l.path)
	if err != nil {
		return nil, "", err
	}
	return dat, "", nil
}

// FileFactory returns a function factory for filesystem loaders.
func FileFactory(path string) Loader {
	return &file{path: path}
}
