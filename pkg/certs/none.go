// The None Provider
package certs

import (
	"time"
)

type NoneProvider struct {
}

func (p *NoneProvider) Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string, ssl string) (keypair KeyPair, certError error) {
	return KeyPair{
		Cert:   []byte{},
		Key:    []byte{},
		Expiry: time.Now(),
	}, nil
}

func (p *NoneProvider) Deprovision(host string) error {
	return nil
}
