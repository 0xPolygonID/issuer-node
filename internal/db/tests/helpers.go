package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func JSONBody(t *testing.T, d any) io.Reader {
	body := &bytes.Buffer{}
	j, err := json.Marshal(d)
	require.NoError(t, err)
	_, err = body.Write(j)
	require.NoError(t, err)
	return body
}
