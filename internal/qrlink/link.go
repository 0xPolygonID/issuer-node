package qrlink

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

const (
	requestURI           = "%s/v2/qr-store?id=%s"
	requestURIWithIssuer = "%s/v2/qr-store?id=%s&issuer=%s"
)

// NewDeepLink creates a deep link
// If issuerDID is nil, it will return a deep link without the issuer DID for backward compatibility
func NewDeepLink(hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		return fmt.Sprintf("iden3comm://?request_uri=%s", url.QueryEscape(fmt.Sprintf(requestURIWithIssuer, hostURL, id.String(), issuerDID.String())))
	}
	return fmt.Sprintf("iden3comm://?request_uri=%s", url.QueryEscape(fmt.Sprintf(requestURI, hostURL, id.String())))
}

// NewUniversal creates a universal link
// If issuerDID is nil, it will return a universal link without the issuer DID for backward compatibility
func NewUniversal(uLinkBaseUrl string, hostURL string, id uuid.UUID, issuerDID *w3c.DID) string {
	if issuerDID != nil {
		requestUri := fmt.Sprintf(requestURIWithIssuer, hostURL, id.String(), issuerDID.String())
		return fmt.Sprintf("%s#request_uri=%s", uLinkBaseUrl, url.QueryEscape(requestUri))
	}
	return fmt.Sprintf("%s#request_uri=%s", uLinkBaseUrl, url.QueryEscape(fmt.Sprintf(requestURI, hostURL, id.String())))
}
