package loader

import (
	"context"
	"net/url"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// MultiProtocol will return a loader for the given url that can be http or ipfs
func MultiProtocol(httpFactory Factory, ipfsFactory Factory, _url string) Loader {
	schemaURL, err := url.Parse(_url)
	if err != nil {
		log.Error(context.Background(), "ipfs factory error", "err", err)
		return nil
	}
	if schemaURL.Scheme == "http" || schemaURL.Scheme == "https" {
		return httpFactory(_url)
	}
	if schemaURL.Scheme == "ipfs" {
		return ipfsFactory(_url)
	}
	log.Error(context.Background(), "unknown package_manager", "url", _url)
	return nil
}

// MultiProtocolFactory will return a factory for the given url that can be http or ipfs
func MultiProtocolFactory(IPFSGateway string) Factory {
	return func(url string) Loader {
		return MultiProtocol(HTTPFactory, IPFSFactory(IPFSGateway, url), url)
	}
}
