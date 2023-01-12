package suite

import (
	"context"
	"errors"
)

// ErrSignerNotDefined is returned when Sign() is called but signer option is not defined.
var ErrSignerNotDefined = errors.New("signer is not defined")

// ErrVerifierNotDefined is returned when Verify() is called but verifier option is not defined.
var ErrVerifierNotDefined = errors.New("verifier is not defined")

// Suite defines general signature suite structure.
type Suite struct {
	Signer   Signer
	Verifier Verifier
}

// Signer is an interface for instances with signing function
type Signer interface {
	// Sign will sign document and return signature
	Sign(ctx context.Context, ata []byte) ([]byte, error)
}

// Opt returns configuration options for cryptographic suite
type Opt func(opts *Suite)

// Verifier is interface to support verify signature function
type Verifier interface {
	// Verify will verify a signature.
	Verify(pubKeyValue []byte, claim, signature []byte) error
}

// WithSigner return new options
func WithSigner(s Signer) Opt {
	return func(opts *Suite) {
		opts.Signer = s
	}
}

// WithVerifier defines a verifier for the Signature Suite.
func WithVerifier(v Verifier) Opt {
	return func(opts *Suite) {
		opts.Verifier = v
	}
}

// InitSuiteOptions initializes signature suite with options.
func InitSuiteOptions(suite *Suite, opts ...Opt) *Suite {
	for _, opt := range opts {
		opt(suite)
	}
	return suite
}

// Verify will verify a signature.
func (s *Suite) Verify(pubKeyValue, data, signature []byte) error {
	if s.Verifier == nil {
		return ErrVerifierNotDefined
	}
	return s.Verifier.Verify(pubKeyValue, data, signature)
}

// Sign will sign input data.
func (s *Suite) Sign(ctx context.Context, data []byte) ([]byte, error) {
	if s.Signer == nil {
		return nil, ErrSignerNotDefined
	}
	return s.Signer.Sign(ctx, data)
}
