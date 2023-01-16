package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	if os.Getenv("TEST_MODE") == "GA" {
		t.Skip("SKIPPED")
	}

	configuration, err := Load("")
	assert.NoError(t, err)
	assert.Equal(t, 3001, configuration.ServerPort)
}
