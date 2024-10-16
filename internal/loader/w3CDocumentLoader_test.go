package loader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestW3CDocumentLoader_LoadDocument(t *testing.T) {
	w3cLoader := NewW3CDocumentLoader(nil, "https://ipfs.io", false)
	doc, err := w3cLoader.LoadDocument(W3CCredential2018ContextURL)
	require.NoError(t, err)

	m, ok := doc.Document.(map[string]interface{})
	require.True(t, ok)
	context, ok := m["@context"]
	require.True(t, ok)
	require.NotNil(t, context)
}
