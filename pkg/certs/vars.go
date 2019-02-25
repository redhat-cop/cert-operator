package certs

import (
	"encoding/json"
	"fmt"
	"github.com/Venafi/vcert"
	"github.com/Venafi/vcert/pkg/endpoint"
	"log"
	"os"
)

var tppConfig = &vcert.Config{
		ConnectorType: endpoint.ConnectorTypeTPP,
		BaseUrl:       os.Getenv("VENAFI_API_URL"),
		Credentials: &endpoint.Authentication{
			User:     os.Getenv("VENAFI_USER_NAME"),
			Password: os.Getenv("VENAFI_PASSWORD")},
		Zone: os.Getenv("VENAFI_CERT_ZONE"),
}

var pp = func(a interface{}) {
	b, err := json.MarshalIndent(a, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	log.Println(string(b))
}