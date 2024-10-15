package services

import (
	"github.com/iden3/iden3comm/v2"
)

// MediaTypeManager manages the list of allowed media types for the package_manager message type
// if strictMode is true, then all messages that do not exist in the allowed list will be rejected
type MediaTypeManager struct {
	enabled   bool
	allowList map[iden3comm.ProtocolMessage][]string
}

// NewMediaTypeManager create instance of MediaTypeManager
func NewMediaTypeManager(allowList map[iden3comm.ProtocolMessage][]string, enabled bool) *MediaTypeManager {
	return &MediaTypeManager{
		enabled:   enabled,
		allowList: allowList,
	}
}

// AllowMediaType check if the package_manager message supports the mediaType type
func (m *MediaTypeManager) AllowMediaType(protoclMessage iden3comm.ProtocolMessage, mediaType iden3comm.MediaType) bool {
	if !m.enabled {
		return true
	}

	al, ok := m.allowList[protoclMessage]
	if !ok {
		return false
	}
	for _, v := range al {
		if v == "*" || v == string(mediaType) {
			return true
		}
	}
	return false
}
