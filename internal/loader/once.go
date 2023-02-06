package loader

import "context"

type once struct {
	loader    Loader
	loaded    bool
	schema    []byte
	extension string
	err       error
}

// Load satisfies Loader interface. It call o.loader only the first time and stores the response. Next calls will use
// previous data
func (o *once) Load(ctx context.Context) (schema []byte, extension string, err error) {
	if !o.loaded {
		o.schema, o.extension, o.err = o.loader.Load(ctx)
		if o.err == nil {
			o.loaded = true
		}
	}
	return o.schema, o.extension, o.err
}

// Once returns a Loader that calls the internal loader l only once, storing the response in memory
func Once(l Loader) Loader {
	return &once{
		loader: l,
	}
}

// OnceFactory returns a factory loader that returns "Once loaders". Loaders that load the file only once
// Once only caches responses. f is the underliying factory loader that will perform the file loading
func OnceFactory(f Factory) Factory {
	return func(url string) Loader {
		return Once(f(url))
	}
}
