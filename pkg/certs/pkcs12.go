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

// CovertToPKCS12 Takes in a crypto private key, x509 certificate, x509 ca chain, password to open the P12 file and returns it as a byte array
func CovertToPKCS12(privateKey []byte, certificate []byte, caCerts [][]byte, password string) ([]byte, error) {
	// convert private key to crypto private key
	privateCryptoKey, err := parsePrivateKey(privateKey)

	// check for error and return if faliure
	if err != nil {
		return nil, err
	}

	// convert certificate to x509
	publicX509, err := parseCert(certificate)
	if err != nil {
		return nil, err
	}

	// convert all ca certs and populate new entry
	caX509Certs := make([]*x509.Certificate, 0, len(caCerts))
	for i, caCert := range caCerts {
		caX509Certs[i], err = parseCert(caCert)

		if err != nil {
			return nil, err
		}
	}

	return ConvertToPKCS12(privateCryptoKey, publicX509, caX509Certs, password)
}

// ConvertToPKCS12 Convert crypto private key, x509 certificate, x509 ca certificates, and password to a p12 password protected file
func ConvertToPKCS12(privateKey crypto.PrivateKey, certificate *x509.Certificate, caCerts []*x509.Certificate, password string) ([]byte, error) {
	return pkcs12.Encode(rand.Reader, privateKey, certificate, caCerts, password)
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

// parse a public cert this can include CA certs and return it as a x509.Certificate
func parseCert(cert []byte) (*x509.Certificate, error) {
	x509Cert, err := x509.ParseCertificate(cert)

	return x509Cert, err
}
