package qrlink

import (
	"testing"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUniversal(t *testing.T) {
	baseURL := "https://wallet-dev.privado.id/"
	hostURL := "https://issuer-node-core-api-testing.privado.id"
	id, err := uuid.Parse("1f209581-ab1d-426d-88d9-2b545bdb851d")
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID("did:iden3:polygon:amoy:x7xjFDkoCW7MSQUZQwrXhyU5HqQ8npzEdAvHmBjqx")
	require.NoError(t, err)
	expected := "https://wallet-dev.privado.id/#request_uri=https%3A%2F%2Fissuer-node-core-api-testing.privado.id%2Fpublic%2Fv2%2Fqr-store%3Fid%3D1f209581-ab1d-426d-88d9-2b545bdb851d%26issuer%3Ddid%3Aiden3%3Apolygon%3Aamoy%3Ax7xjFDkoCW7MSQUZQwrXhyU5HqQ8npzEdAvHmBjqx"
	got := NewUniversal(baseURL, hostURL, id, issuerDID)
	assert.Equal(t, expected, got)
}

func TestDeepLink(t *testing.T) {
	hostURL := "https://issuer-node-core-api-testing.privado.id"
	id, err := uuid.Parse("1f209581-ab1d-426d-88d9-2b545bdb851d")
	require.NoError(t, err)
	issuerDID, err := w3c.ParseDID("did:iden3:polygon:amoy:x7xjFDkoCW7MSQUZQwrXhyU5HqQ8npzEdAvHmBjqx")
	require.NoError(t, err)
	expected := "iden3comm://?request_uri=https%3A%2F%2Fissuer-node-core-api-testing.privado.id%2Fpublic%2Fv2%2Fqr-store%3Fid%3D1f209581-ab1d-426d-88d9-2b545bdb851d%26issuer%3Ddid%3Aiden3%3Apolygon%3Aamoy%3Ax7xjFDkoCW7MSQUZQwrXhyU5HqQ8npzEdAvHmBjqx"
	got := NewDeepLink(hostURL, id, issuerDID)
	assert.Equal(t, expected, got)
}
