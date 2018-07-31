package certs

import "time"

type Provider interface {
	Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string) (KeyPair, error)
	Deprovision(host string) error
}

type ProviderConfig struct {
	Kind string `json:"kind"`
}

type KeyPair struct {
	Cert   []byte
	Key    []byte
	Expiry time.Time
}
