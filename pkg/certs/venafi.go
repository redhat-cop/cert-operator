package certs

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/Venafi/vcert"
	"github.com/Venafi/vcert/pkg/certificate"
	t "log"
	"net/http"
	"time"
)

type VenafiProvider struct {
}

func (p *VenafiProvider) Provision(host string, validFrom string, validFor time.Duration, isCA bool, rsaBits int, ecdsaCurve string) (keypair KeyPair, certError error) {
	
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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

	//
	// 0. get client instance based on connection config
	//
	c, err := vcert.NewClient(tppConfig)
	if err != nil {
		t.Fatalf("could not connect to endpoint: %s", err)
	}

	//
	// 1.1. compose request object
	//
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
		KeyPassword:    "newPassw0rd!",
	}

	//
	// 1.2. generate private key and certificate request (CSR) based on request's options
	//
	err = c.GenerateRequest(nil, enrollReq)
	if err != nil {
		t.Fatalf("could not generate certificate request: %s", err)
	}

	//
	// 1.3. submit certificate request, get request ID as a response
	//
	requestID, err := c.RequestCertificate(enrollReq, "")
	if err != nil {
		t.Fatalf("could not submit certificate request: %s", err)
	}
	t.Printf("Successfully submitted certificate request. Will pickup certificate by ID %s", requestID)

	//
	// 1.4. retrieve certificate using request ID obtained on previous step, get PEM collection as a response
	//
	pickupReq := &certificate.Request{
		PickupID: requestID,
		Timeout:  180 * time.Second,
	}
	pcc, err := c.RetrieveCertificate(pickupReq)
	if err != nil {
		t.Fatalf("could not retrieve certificate using requestId %s: %s", requestID, err)
	}

	//
	// 1.5. (optional) add certificate's private key to PEM collection
	//
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