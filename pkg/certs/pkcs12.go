package certs

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"

	"software.sslmate.com/src/go-pkcs12"
)

// ConvertToPKCS12 Takes in a crypto private key, x509 certificate, x509 ca chain, password to open the P12 file and returns it as a byte array
func ConvertToPKCS12(privateKey []byte, certificate []byte, caCerts [][]byte, password string) ([]byte, error) {
	// convert private key to crypto private key
	privateCryptoKey, err := parsePrivateKey(privateKey)

	// check for error and return if faliure
	if err != nil {
		return nil, err
	}

	// convert certificate to x509
	publicX509, err := x509.ParseCertificate(certificate)
	if err != nil {
		return nil, err
	}

	// convert all ca certs and populate new entry
	caX509Certs := make([]*x509.Certificate, len(caCerts))
	for i, caCert := range caCerts {
		// TODO would like to set the array directly if possible
		x509CaCert, err := x509.ParseCertificate(caCert)
		caX509Certs[i] = x509CaCert

		if err != nil {
			return nil, err
		}
	}
	return pkcs12.Encode(rand.Reader, privateCryptoKey, publicX509, caX509Certs, password)
}

// Convert the bytes of a private key to a crypto.PrivateKey
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	// try to parse as a PKCS1 private key
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	// try to parse as a PCKSC8 private key
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("crypto/tls: found unknown private key type in PKCS#8 wrapping")
		}
	}

	// try to parse as an EC private key
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	// was not able to parse so return appropriate error
	return nil, errors.New("crypto/tls: failed to parse private key")
}
