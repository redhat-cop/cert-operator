package certs

import (
	"fmt"
	"os"
	"time"

	"github.com/Venafi/govcert"
	vcert "github.com/Venafi/govcert/embedded"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/sirupsen/logrus"
)

type VenafiProvider struct {
}

func (p *VenafiProvider) Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string) (keypair KeyPair, certError error) {

	if len(host) == 0 {
		return KeyPair{}, NewErrBadHost("host cannot be empty")
	}
	logrus.Infof("Creating Venafi certificate for :" + host)
	logrus.Infof("Venafi Zone: " + os.Getenv("VENAFI_CERT_ZONE"))

	var notBefore time.Time
	var err error
	if len(validFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			return KeyPair{}, NewCertError("Failed to parse creation date: " + err.Error())
		}
	}

	notAfter := notBefore.Add(validFor)

	username := os.Getenv("VENAFI_USER_NAME")
	password := os.Getenv("VENAFI_PASSWORD")
	url := os.Getenv("VENAFI_API_URL")
	venafi := vcert.NewClientTPP(username, password, url)
	enrollreq := &govcert.EnrollReq{
		CommonName: host,
		Zone:       os.Getenv("VENAFI_CERT_ZONE")}

	resp, err := venafi.Do(enrollreq)
	if err != nil {
		return KeyPair{}, NewCertError("Failed to create certificate: " + err.Error())
	}
	respmap, err := resp.JSONBody()
	if err != nil {
		return KeyPair{}, NewCertError("Failed to create certificate: " + err.Error())
	}
	var cert, key []byte

	key = []byte(respmap["PrivateKey"].(string))

	if !resp.Pending() {
		cert = []byte(respmap["Certificate"].(string))

		if err != nil {
			return KeyPair{}, NewCertError("Failed to create certificate: " + err.Error())
		}
		return KeyPair{
			cert,
			key,
			notAfter}, nil
	}
	id, err := resp.RequestID()
	if err != nil {
		return KeyPair{}, NewCertError("Failed to create certificate: " + err.Error())
	}
	pickupreq := &govcert.PickupReq{
		PickupID: id,
	}
	retryerr := resource.Retry(time.Duration(300)*time.Second, func() *resource.RetryError {
		resp, err = venafi.Do(pickupreq)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if resp.Pending() {
			return resource.RetryableError(fmt.Errorf("Certificate Issue pending"))
		}

		return nil
	})
	if retryerr != nil {
		return KeyPair{}, NewCertError("Failed to create certificate: " + retryerr.Error())
	}

	respmap, err = resp.JSONBody()
	if err != nil {
		return KeyPair{}, NewCertError("Failed to create certificate: " + err.Error())
	}
	cert = []byte(respmap["Certificate"].(string))

	return KeyPair{
		cert,
		key,
		notAfter}, nil
}
func (p *VenafiProvider) Deprovision(host string) error {
	return nil
}
