package certs

import "time"

type Provider interface {
	Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string) (cert []byte, key []byte, err error)
	Deprovision(host string) error
}
