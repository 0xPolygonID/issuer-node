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
		strictMode            bool
	}
	testcases := []testcase{
		{
			name: "strictMode = true. Protocol message not in the allow list",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.RevocationStatusRequestMessageType: {"*"},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              false,
			strictMode:            true,
		},
		{
			name: "strictMode = false. Protocol message not in the allow list",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.RevocationStatusRequestMessageType: {"*"},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
			strictMode:            false,
		},
		{
			name: "Protocol message on the allow list with '*'",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {"*"},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              true,
			strictMode:            true,
		},
		{
			name: "Protocol message on the allow list with allow media type",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {string(packers.MediaTypeZKPMessage)},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
			strictMode:            true,
		},
		{
			name: "Protocol message on the allow list with NOT allow media type",
			allowList: map[iden3comm.ProtocolMessage][]string{
				protocol.CredentialFetchRequestMessageType: {string(packers.MediaTypeZKPMessage)},
			},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              false,
			strictMode:            true,
		},
		{
			name:                  "strictMode = true. Empty allow list",
			allowList:             map[iden3comm.ProtocolMessage][]string{},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              false,
			strictMode:            true,
		},
		{
			name:                  "strictMode = false. Empty allow list",
			allowList:             map[iden3comm.ProtocolMessage][]string{},
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              true,
			strictMode:            false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mdm := services.NewMediaTypeManager(
				tt.allowList, tt.strictMode,
			)
			actual := mdm.AllowMediaType(
				tt.targetProtocolMessage, tt.targetMediatype,
			)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestMediatypeManager_DefaultPacker(t *testing.T) {
	type testcase struct {
		name                  string
		targetProtocolMessage iden3comm.ProtocolMessage
		targetMediatype       iden3comm.MediaType
		expected              bool
	}
	testcases := []testcase{
		{
			name:                  "call CredentialFetchRequestMessageType with MediaTypePlainMessage",
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              false,
		},
		{
			name:                  "call CredentialFetchRequestMessageType with MediaTypeZKPMessage",
			targetProtocolMessage: protocol.CredentialFetchRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
		},
		{
			name:                  "call RevocationStatusRequestMessageType with MediaTypePlainMessage",
			targetProtocolMessage: protocol.RevocationStatusRequestMessageType,
			targetMediatype:       packers.MediaTypePlainMessage,
			expected:              true,
		},
		{
			name:                  "call RevocationStatusRequestMessageType with MediaTypeZKPMessage",
			targetProtocolMessage: protocol.RevocationStatusRequestMessageType,
			targetMediatype:       packers.MediaTypeZKPMessage,
			expected:              true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			actual := services.DefaultMediaTypeManager.AllowMediaType(tt.targetProtocolMessage, tt.targetMediatype)
			require.Equal(t, tt.expected, actual)
		})
	}
}
