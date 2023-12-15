package loader

import (
	"context"
	"errors"
	"net/url"

	"github.com/iden3/go-schema-processor/loaders"

	"github.com/polygonid/issuer-node/internal/log"
)

// IPFS loader
func IPFS(IPFSGateway string, _url string) Loader {
	cid, err := ipfsCID(_url)
	if err != nil {
		log.Error(context.Background(), "ipfs factory error", "err", err)
		return nil
	}
	return &loaders.IPFS{URL: IPFSGateway, CID: cid}
}

// IPFSFactory returns an ipfs  loader
func IPFSFactory(IPFSGateway string, _url string) Factory {
	return func(url string) Loader {
		return IPFS(IPFSGateway, _url)
	}
}

func ipfsCID(ipfsURL string) (cid string, err error) {
	schemaURL, err := url.Parse(ipfsURL)
	if err != nil {
		return ipfsURL, err
	}
	if schemaURL.Scheme == "ipfs" {
		return schemaURL.Host, nil
	}
	return "", errors.New("invalid ipfs url")
}
