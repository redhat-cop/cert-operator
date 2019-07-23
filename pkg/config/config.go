/*
Sample config file

notifiers:
- name: myslacknotifier
  kind: slack
  integration_url: https://hooks.slack.com/services/service_id/auth-token
*/

package config

import (
	"encoding/json"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/flag"
	"os"

	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/memory"
	"github.com/redhat-cop/cert-operator/pkg/certs"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("Config")

type Config struct {
	//	Notifiers []notifier.Notifier  `json:"notifiers"`
	Provider certs.ProviderConfig `json:"provider"`
	General  GeneralConfig        `json:"general"`
}

type GeneralConfig struct {
	Annotations AnnotationConfig `json:"annotations"`
}

type AnnotationConfig struct {
	Status        string `json:"status"`
	StatusReason  string `json:"status-reason"`
	Expiry        string `json:"expiry"`
	Format        string `json:"format"`
	NeedCertValue string `json:"need-cert-value"`
	PemFormat     string `json:"pem-format-value"`
	Pkcs12Format  string `json:"pkcs12-format-value"`
}

const (
	defaultConfigFile = "/etc/cert-operator/config.yaml"
	defaultConfig     = `
  {
    "general": {
      "annotations": {
        "status": "openshift.io/cert-ctl-status",
        "status-reason": "openshift.io/cert-ctl-status-reason",
        "expiry": "openshift.io/cert-ctl-expires",
        "format": "openshift.io/cert-ctl-format",
        "need-cert-value": "new",
        "pem-format-value": "PEM",
        "pkcs12-format-value": "PKCS12"
      }
    },
    "provider": {
      "kind": "self-signed",
      "ssl": "false"
    }
  }`
)

func NewConfig() Config {

	tmpConfig := config.NewConfig()

	data := []byte(defaultConfig)

	memorySource := memory.NewSource(
		memory.WithData(data),
	)

	log.Info("Using default configuration.")
	_ = tmpConfig.Load(
		memorySource,
	)

	configFile := getConfigFile()
	if fileExist(configFile) {
		log.Info("Loading config file from ", configFile)

		_ = tmpConfig.Load(file.NewSource(
			file.WithPath(configFile),
		))
	}

	_ = tmpConfig.Load(
		env.NewSource(),
		flag.NewSource(),
	)
	var conf Config

	err := tmpConfig.Scan(&conf)

	if err != nil {
		log.Error(err, "Could not load configurations properly.")
	}

	return conf
}

func getConfigFile() (configFile string) {
	if value, ok := os.LookupEnv("CERT_OP_CONFIG"); ok {
		return value
	}
	return defaultConfigFile
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (c *Config) String() string {
	out, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}
