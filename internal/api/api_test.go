package api

import (
	"context"
	"testing"

	"github.com/deepmap/oapi-codegen/examples/petstore-expanded/chi/api"
	"github.com/stretchr/testify/assert"
)

// TestValid verifies that the api spec can ve validated by github.com/getkin/kin-openapi
func TestValid(t *testing.T) {
	spec, err := api.GetSwagger()
	assert.NoError(t, err)
	assert.NoError(t, spec.Validate(context.Background()))
}
