package services

type claimConfig struct {
	RHSEnabled bool
	RHSUrl     string
	Host       string
}

type claimsConfigFunc func(c *claimConfig)

// WithReverseHashEnabled enables the claim service to return claims with reverse hashes
func WithReverseHashEnabled(enabled bool, host string) claimsConfigFunc {
	return func(c *claimConfig) {
		c.RHSEnabled = enabled
		c.RHSUrl = host
	}
}

// WithHost set the host to be used for creating ids
func WithHost(h string) claimsConfigFunc {
	return func(c *claimConfig) {
		c.Host = h
	}
}
