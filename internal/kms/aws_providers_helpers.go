package kms

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/polygonid/sh-id-platform/internal/log"
)

//type publicKeyInfo struct {
//	Raw       asn1.RawContent
//	Algorithm pkix.AlgorithmIdentifier
//	PublicKey asn1.BitString
//}

//type ECDSASignature struct {
//	R, S *big.Int
//}

type asn1EcSig struct {
	R asn1.RawValue
	S asn1.RawValue
}

type asn1EcPublicKey struct {
	EcPublicKeyInfo asn1EcPublicKeyInfo
	PublicKey       asn1.BitString
}

type asn1EcPublicKeyInfo struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

// DecodeAWSETHPubKey decodes the public key from the AWS KMS response.
func DecodeAWSETHPubKey(_ context.Context, key []byte) (*ecdsa.PublicKey, error) {
	var asn1pubk asn1EcPublicKey
	_, err := asn1.Unmarshal(key, &asn1pubk)
	if err != nil {
		return nil, err
	}
	pubkey, err := crypto.UnmarshalPubkey(asn1pubk.PublicKey.Bytes)
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

// DecodeAWSETHSig decodes the signature from the AWS KMS response
func DecodeAWSETHSig(ctx context.Context, signature []byte, pubKeyBytes []byte, data []byte) ([]byte, error) {
	const secp256k1HalfNNumber = 2
	var sigAsn1 asn1EcSig
	// nolint:all
	var secp256k1N = crypto.S256().Params().N
	var secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(secp256k1HalfNNumber))

	_, err := asn1.Unmarshal(signature, &sigAsn1)
	if err != nil {
		return nil, err
	}
	sBigInt := new(big.Int).SetBytes(sigAsn1.S.Bytes)
	if sBigInt.Cmp(secp256k1HalfN) > 0 {
		sigAsn1.S.Bytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
	}

	ethSignature, err := getEthereumSignature(ctx, pubKeyBytes, data, sigAsn1.R.Bytes, sigAsn1.S.Bytes)
	if err != nil {
		return nil, err
	}

	if !verifySignature(pubKeyBytes, data, ethSignature) {
		log.Error(ctx, "signature verification failed")
		return nil, errors.New("signature verification failed")
	}
	return ethSignature, nil
}

func verifySignature(pubKey []byte, message, signature []byte) bool {
	r := signature[:32]
	s := signature[32:64]
	rs := append(r, s...)

	// verify signature
	return crypto.VerifySignature(pubKey, message, rs)
}

// getEthereumSignature reconstructs the signature in the Ethereum format
func getEthereumSignature(ctx context.Context, expectedPublicKeyBytes []byte, data []byte, r []byte, s []byte) ([]byte, error) {
	rsSignature := append(adjustSignatureLength(r), adjustSignatureLength(s)...)
	signature := append(rsSignature, []byte{0}...)
	recoveredPublicKeyBytes, err := crypto.Ecrecover(data, signature)
	if err != nil {
		log.Error(ctx, "failed to recover public key", "err", err)
		return nil, err
	}

	if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
		signature = append(rsSignature, []byte{1}...)
		recoveredPublicKeyBytes, err = crypto.Ecrecover(data, signature)
		if err != nil {
			log.Error(ctx, "failed to recover public key", "err", err)
			return nil, err
		}

		if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
			return nil, errors.New("can not reconstruct public key from sig")
		}
	}

	return signature, nil
}

func adjustSignatureLength(buffer []byte) []byte {
	buffer = bytes.TrimLeft(buffer, "\x00")
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}
	return buffer
}

//func DecodeAWSETHSig(ctx context.Context, key []byte, payload []byte, pk *ecdsa.PublicKey) ([]byte, error) {
//	block, rest := pem.Decode(key)
//	var toUnmarshal []byte
//	if block != nil {
//		toUnmarshal = block.Bytes
//	} else {
//		toUnmarshal = rest
//	}
//
//	sigRS := &ECDSASignature{}
//	_, err := asn1.Unmarshal(toUnmarshal, sigRS)
//	if err != nil {
//		log.Error(ctx, "failed to unmarshal aws public key", "err", err)
//		return nil, err
//	}
//
//	//v := byte(0)
//	//if sigRS.S.Cmp(secp256k1halfN()) == 1 {
//	//	v = 1
//	//}
//	//
//	//signature := append(sigRS.R.Bytes(), sigRS.S.Bytes()...)
//	//signature = append(signature, v)
//
//	signature, err := encodeSignature(sigRS.R.Bytes(), sigRS.S.Bytes())
//	if err != nil {
//		log.Error(ctx, "failed to encode signature", "err", err)
//		return nil, err
//	}
//
//	ok := ecdsa.Verify(pk, payload[:], sigRS.R, sigRS.S)
//	if !ok {
//		log.Error(ctx, "failed to verify signature", "err", err)
//	}
//	return signature, nil
//}
