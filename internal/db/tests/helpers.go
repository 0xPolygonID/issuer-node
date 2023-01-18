package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// JSONBody is a helper function to create an http body suitable to use in calls to http.NewRequest
// You can pass any object and you'll get an io.Reader that returns this object in json format
func JSONBody(t *testing.T, d any) io.Reader {
	body := &bytes.Buffer{}
	j, err := json.Marshal(d)
	require.NoError(t, err)
	_, err = body.Write(j)
	require.NoError(t, err)
	return body
}
