package services_test

import (
	"testing"

	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/stretchr/testify/require"

	"github.com/polygonid/sh-id-platform/internal/core/services"
)

func TestMediatypeManager_AllowList(t *testing.T) {
	type testcase struct {
		name                  string
		allowList             map[iden3comm.ProtocolMessage][]string
		targetProtocolMessage iden3comm.ProtocolMessage
		targetMediatype       iden3comm.MediaType
		expected              bool
		enabled               bool
	}
	testcases := []testcase{
		{
			name: "AllowList enabled. Type in the list",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {string(packers.MediaTypeZKPMessage)},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
			enabled:               true,
		},
		{
			name: "AllowList enabled. Type in the list with wildcard",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {"*"},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
			enabled:               true,
		},
		{
			name: "AllowList enabled. Type not in the list",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.RevocationStatusRequestMessageType: {"*"},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              false,
			enabled:               true,
		},
		{
			name: "AllowList enabled. Type does not exist",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {string(packers.MediaTypePlainMessage)},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              false,
			enabled:               true,
		},
		{
			name: "AllowList disabled. Type does not exist",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {string(packers.MediaTypePlainMessage)},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
			enabled:               false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mdm := services.NewMediaTypeManager(
				tt.allowList, tt.enabled,
			)
			actual := mdm.AllowMediaType(
				tt.targetProtocolMessage, tt.targetMediatype,
			)
			require.Equal(t, tt.expected, actual)
		})
	}
}
