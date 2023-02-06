package loader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type spyLoader struct {
	called int // We will count the number of times the Load function is called
}

func (s *spyLoader) Load(_ context.Context) (schema []byte, extension string, err error) {
	s.called++
	return []byte("this is an schema content"), "extension", nil
}

func TestOnce_Load(t *testing.T) {
	ctx := context.Background()
	spy := &spyLoader{}
	myLoader := OnceFactory(func(url string) Loader { return spy })("http://this/is/an/url")
	assert.Equal(t, spy.called, 0)
	for i := 0; i < 100; i++ {
		schema, ext, err := myLoader.Load(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []byte("this is an schema content"), schema)
		assert.Equal(t, "extension", ext)
		assert.Equal(t, 1, spy.called, "Load function of underlying loader has only been called once")
	}
}
