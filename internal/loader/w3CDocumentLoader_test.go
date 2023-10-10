package loader

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestW3CDocumentLoader_LoadDocument(t *testing.T) {
	w3cLoader := NewW3CDocumentLoader(nil, "https://ipfs.io")
	doc, err := w3cLoader.LoadDocument(W3CCredential2018ContextURL)
	require.NoError(t, err)

	require.NotNil(t, (doc.Document.(map[string]interface{}))["@context"])
}
