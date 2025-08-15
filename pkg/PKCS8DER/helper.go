package PKCS8DER

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var (
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
	oidNamedCurveS256 = asn1.ObjectIdentifier{1, 3, 132, 0, 10}
)

const (
	bitLengthMultiplier = 8 // Number of bits in a byte
)

type pkcs8 struct {
	Version    int
	Algo       pkixAlgorithmIdentifier
	PrivateKey []byte
}

type pkixAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"explicit,tag:1"`
}

// MarshalECPrivateKeyToPKCS8DER serializes an ECDSA private key using the secp256k1 curve
// into an unencrypted PKCS#8 DER-encoded byte slice.
//
// This function is necessary because the Go standard library's x509.MarshalPKCS8PrivateKey
// does not support the secp256k1 curve, which is used in Ethereum and other blockchain applications.
// AWS KMS expects imported asymmetric private keys to be in unencrypted PKCS#8 format,
// including an AlgorithmIdentifier with the secp256k1 OID (1.3.132.0.10).
//
// This custom implementation ensures compatibility with AWS KMS by generating
// a valid ASN.1 structure that includes the ECPrivateKey (RFC 5915) nested within
// the PKCS#8 wrapper, using the correct curve identifiers.
func MarshalECPrivateKeyToPKCS8DER(priv *ecdsa.PrivateKey) ([]byte, error) {
	if priv == nil {
		return nil, errors.New("nil private key")
	}
	if priv.Curve != secp256k1.S256() {
		return nil, errors.New("unsupported curve: only secp256k1 is supported")
	}

	pubKeyBytes := append([]byte{0x04}, priv.X.Bytes()...)
	pubKeyBytes = append(pubKeyBytes, priv.Y.Bytes()...)

	ecKey := ecPrivateKey{
		Version:       1,
		PrivateKey:    priv.D.Bytes(),
		NamedCurveOID: oidNamedCurveS256,
		PublicKey:     asn1.BitString{Bytes: pubKeyBytes, BitLength: len(pubKeyBytes) * bitLengthMultiplier},
	}
	ecKeyDer, err := asn1.Marshal(ecKey)
	if err != nil {
		return nil, err
	}

	pkcs8Key := pkcs8{
		Version: 0,
		Algo: pkixAlgorithmIdentifier{
			Algorithm:  oidPublicKeyECDSA,
			Parameters: oidNamedCurveS256,
		},
		PrivateKey: ecKeyDer,
	}

	return asn1.Marshal(pkcs8Key)
}
