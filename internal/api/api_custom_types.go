package api

import "github.com/iden3/go-iden3-core/v2/w3c"

// Identity represents a DID
type Identity struct {
	w3cDID *w3c.DID
}

// UnmarshalText unmarshals a DID from text
func (c *Identity) UnmarshalText(text []byte) error {
	did, err := w3c.ParseDID(string(text))
	if err != nil {
		return err
	}
	c.w3cDID = did
	return nil
}

func (c *Identity) did() *w3c.DID {
	return c.w3cDID
}
