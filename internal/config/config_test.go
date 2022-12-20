package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	configuration, err := Load("../../")
	assert.NoError(t, err)
	assert.Equal(t, "3001", configuration.ServerPort)
}
