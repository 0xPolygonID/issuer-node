package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyMap(t *testing.T) {
	m1 := map[string]interface{}{
		"hello": "world",
		"foo": map[string]interface{}{
			"bar": 123,
		},
	}

	m2 := CopyMap(m1)

	m1["hello"] = "!world"
	delete(m1, "foo")

	require.Equal(t, map[string]interface{}{"hello": "!world"}, m1)
	require.Equal(t, map[string]interface{}{
		"hello": "world",
		"foo": map[string]interface{}{
			"bar": 123,
		},
	}, m2)
}
