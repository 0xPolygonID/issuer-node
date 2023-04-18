package jsonschema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/polygonid/sh-id-platform/internal/loader"
)

func Test_Attributes(t *testing.T) {
	ctx := context.Background()
	const url = "https://raw.githubusercontent.com/iden3/claim-schema-vocab/main/schemas/json/KYCAgeCredential-v3.json"
	jsonSchema, err := Load(ctx, loader.HTTPFactory(url))
	assert.NoError(t, err)
	attributes, err := jsonSchema.Attributes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(attributes))
}
