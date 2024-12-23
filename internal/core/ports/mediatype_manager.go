package ports

import "github.com/iden3/iden3comm/v2"

// MediaTypeManager - Define the interface for media type manager
type MediaTypeManager interface {
	AllowMediaType(protocolMessage iden3comm.ProtocolMessage, mediaType iden3comm.MediaType) bool
}
