package certs

import (
	"encoding/pem"
	"io/ioutil"
	"testing"

	"software.sslmate.com/src/go-pkcs12"
)

func TestConvertToPKCS12(t *testing.T) {
	// read in test certs
	// setup
	privKeyBytes := readFile("../../test/certs/testCerts/example.com.key")
	certBytes := readFile("../../test/certs/testCerts/example.com.crt")
	// rootCABytes := readFile("../../test/certs/testCerts/rootCA.crt")
	rootCAS := [][]byte{}

	// act
	pkcs12Byte, err := ConvertToPKCS12(privKeyBytes, certBytes, rootCAS, "secret")

	// assert
	if err != nil {
		t.Fatal(err)
	}

	// decode and read cert
	privKey, cert, err := pkcs12.Decode(pkcs12Byte, "secret")

	if err != nil {
		t.Fatal(err)
	}

	if privKey == nil {
		t.Fatal("private key is nil")
	}

	if cert.Issuer.CommonName != "TESTING_CA" {
		t.Fatal("invalid issuer")
	}

	if cert.Subject.CommonName != "test.example.com" {
		t.Fatal("invalid domain name")
	}
}

func readFile(file string) []byte {
	var f = file
	r, _ := ioutil.ReadFile(f)
	block, _ := pem.Decode(r)

	return block.Bytes
}
