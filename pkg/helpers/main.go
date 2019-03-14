package helpers

import (
	"context"
	"time"

	"github.com/redhat-cop/cert-operator/pkg/certs"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TimeFormat = "Jan 2 15:04:05 2006"
)

func Apply(c client.Client, object runtime.Object) error {
	err := c.Create(context.TODO(), object)
	if err != nil {
		err = c.Update(context.TODO(), object)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func GetCert(host string, provider certs.Provider, ssl string) (certs.KeyPair, error) {
	oneYear, timeErr := time.ParseDuration("8760h")
	if timeErr != nil {
		return certs.KeyPair{}, timeErr
	}

	// Retreive cert from provider
	keyPair, err := provider.Provision(
		host,
		time.Now().Format(TimeFormat),
		oneYear, false, 2048, "", ssl)
	if err != nil {
		return certs.KeyPair{}, err
	}
	return keyPair, nil
}
