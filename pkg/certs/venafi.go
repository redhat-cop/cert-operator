package certs

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/Venafi/vcert"
	"github.com/Venafi/vcert/pkg/certificate"
	t "log"
	"net/http"
	"time"
	"io/ioutil"
	"os"
	"github.com/Venafi/vcert/pkg/endpoint"
	"encoding/json"
	"fmt"
)

type VenafiProvider struct {
}

/*
 The Provision function follows the example provided by Venafi.
 https://github.com/Venafi/vcert/blob/master/example/main.go
*/

func (p *VenafiProvider) Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string, ssl string) (keypair KeyPair, certError error) {

	if len(host) == 0 {
		return KeyPair{}, NewErrBadHost("host cannot be empty")
	}

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

	var tppConfig = &vcert.Config{}

	if ssl == "on" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{}

		trustBundle, err := ioutil.ReadFile(os.Getenv("VENAFI_CERT_PATH"))
		if err != nil {
			NewCertError("trust was not found in path")
		}
		trustBundlePEM := string(trustBundle)

		tppConfig = &vcert.Config{
				ConnectorType: endpoint.ConnectorTypeTPP,
				BaseUrl:       os.Getenv("VENAFI_API_URL"),
				ConnectionTrust: trustBundlePEM,
				Credentials: &endpoint.Authentication{
					User:     os.Getenv("VENAFI_USER_NAME"),
					Password: os.Getenv("VENAFI_PASSWORD")},
				Zone: os.Getenv("VENAFI_CERT_ZONE"),
		}

	} else {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

		tppConfig = &vcert.Config{
				ConnectorType: endpoint.ConnectorTypeTPP,
				BaseUrl:       os.Getenv("VENAFI_API_URL"),
				Credentials: &endpoint.Authentication{
					User:     os.Getenv("VENAFI_USER_NAME"),
					Password: os.Getenv("VENAFI_PASSWORD")},
				Zone: os.Getenv("VENAFI_CERT_ZONE"),
		}
	}

	c, err := vcert.NewClient(tppConfig)
	if err != nil {
		t.Fatalf("could not connect to endpoint: %s", err)
	}

	enrollReq := &certificate.Request{
		Subject: pkix.Name{
			CommonName:         host,
			Organization:       []string{"Venafi.com"},
			OrganizationalUnit: []string{"Integration Team"},
			Locality:           []string{"Salt Lake"},
			Province:           []string{"Salt Lake"},
			Country:            []string{"US"},
		},
		DNSNames:       []string{host},
		CsrOrigin:      certificate.LocalGeneratedCSR,
		KeyType:        certificate.KeyTypeRSA,
		KeyLength:      2048,
		ChainOption:    certificate.ChainOptionRootLast,
	}

	err = c.GenerateRequest(nil, enrollReq)
	if err != nil {
		t.Fatalf("could not generate certificate request: %s", err)
	}

	requestID, err := c.RequestCertificate(enrollReq, "")
	if err != nil {
		t.Fatalf("could not submit certificate request: %s", err)
	}
	t.Printf("Successfully submitted certificate request. Will pickup certificate by ID %s", requestID)

	pickupReq := &certificate.Request{
		PickupID: requestID,
		Timeout:  180 * time.Second,
	}
	pcc, err := c.RetrieveCertificate(pickupReq)
	if err != nil {
		t.Fatalf("could not retrieve certificate using requestId %s: %s", requestID, err)
	}

	pcc.AddPrivateKey(enrollReq.PrivateKey, []byte(enrollReq.KeyPassword))

	t.Printf("Successfully picked up certificate for %s", host)
	pp(pcc)

	var cert = []byte(pcc.Certificate)
    var privateKey = []byte(pcc.PrivateKey)

	return KeyPair{
		cert,
		privateKey,
		notAfter}, nil
}

func (p *VenafiProvider) Deprovision(host string) error {
	return nil
}

var pp = func(a interface{}) {
	b, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	t.Println(string(b))
} 
