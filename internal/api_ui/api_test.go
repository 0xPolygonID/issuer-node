package api_ui

import (
	"context"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValid verifies that the api spec can ve validated by github.com/getkin/kin-openapi
func TestValid(t *testing.T) {
	file, err := os.ReadFile("../../api_ui/api.yaml")
	require.NoError(t, err)
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(file)
	require.NoError(t, err)

	assert.NoError(t, spec.Validate(context.Background()))
}
