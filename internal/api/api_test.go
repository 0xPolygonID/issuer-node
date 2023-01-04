package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValid verifies that the api spec can ve validated by github.com/getkin/kin-openapi
func TestValid(t *testing.T) {
	spec, err := GetSwagger()
	assert.NoError(t, err)
	assert.NoError(t, spec.Validate(context.Background()))
}
