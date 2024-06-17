package services

import (
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
)

// DefaultMediaTypeManager default media type manager
var DefaultMediaTypeManager = NewMediaTypeManager(
	map[iden3comm.ProtocolMessage][]string{
		protocol.CredentialFetchRequestMessageType:  {string(packers.MediaTypeZKPMessage)},
		protocol.RevocationStatusRequestMessageType: {"*"},
	},
	true,
)

// MediaTypeManager manages the list of allowed media types for the protocol message type
// if strictMode is true, then all messages that do not exist in the allowed list will be rejected
type MediaTypeManager struct {
	strictMode bool
	allowList  map[iden3comm.ProtocolMessage][]string
}

// NewMediaTypeManager create instance of MediaTypeManager
func NewMediaTypeManager(allowList map[iden3comm.ProtocolMessage][]string, strictMode bool) MediaTypeManager {
	return MediaTypeManager{
		strictMode: strictMode,
		allowList:  allowList,
	}
}

// AllowMediaType check if the protocol message supports the mediaType type
func (m *MediaTypeManager) AllowMediaType(
	protoclMessage iden3comm.ProtocolMessage,
	mediaType iden3comm.MediaType,
) bool {
	al, ok := m.allowList[protoclMessage]
	if !ok {
		return !m.strictMode
	}
	for _, v := range al {
		if v == "*" || v == string(mediaType) {
			return true
		}
	}
	return false
}
