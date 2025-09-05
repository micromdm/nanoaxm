// Package cyrptoutil provides cryptographic utilities.
package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ECPrivateKeyFromPEM attempts to parse the PEM EC private key.
// This key is distributed from the Apple Business Manager or
// Apple School Manager portal.
func ECPrivateKeyFromPEM(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)

	// TODO: parse it correctly first, fallback to weirdness
	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		// Apple seems to have distributed the private key as PKCS#8
		// yet uses the "EC PRIVATE KEY" marker on the PEM.
		p8Key, p8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if p8Err != nil {
			return nil, fmt.Errorf("after parsing EC private key: %v; parsing PKCS8 private key: %w", err, p8Err)
		}

		var ok bool
		key, ok = p8Key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("parsed PKCS8 is not ecdsa.PrivateKey")
		}
	}

	return key, nil
}
