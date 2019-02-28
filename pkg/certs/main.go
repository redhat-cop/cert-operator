package certs

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"
)

type Provider interface {
	Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string, ssl string) (KeyPair, error)
	Deprovision(host string) error
}

type ProviderConfig struct {
	Kind string `json:"kind"`
	Ssl string  `json:"ssl"`
}

type KeyPair struct {
	Cert   []byte
	Key    []byte
	Expiry time.Time
}

// Shared functions
func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, NewCertError("Unable to marshal ECDSA private key: " + err.Error())
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, NewCertError("Ran out of possible options for PEM Block")
	}
}
