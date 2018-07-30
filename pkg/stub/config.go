/*
Sample config file

notifiers:
- name: myslacknotifier
  kind: slack
  integration_url: https://hooks.slack.com/services/service_id/auth-token
*/

package stub

import (
	"os"

	config "github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
	"github.com/micro/go-config/source/memory"
)

type Config struct {
	Notifiers []NotifierConfig `json:"notifiers"`
	Provider  ProviderConfig   `json:"provider"`
}

type NotifierConfig struct {
	name string `json:"name"`
}

type ProviderConfig struct {
	Kind string `json:"kind"`
}

const (
	defaultConfigFile = "/etc/cert-operator/config.yml"
	defaultProvider   = "self-signed"
	defaultConfig     = `
  {
    "provider": {
      "kind": "self-signed"
    },
    "notifiers": [
      {
        "kind": "log"
      }
    ]
  }`
)

func NewConfig() Config {

	tmpConfig := config.NewConfig()

	data := []byte(defaultConfig)

	memorySource := memory.NewSource(
		memory.WithData(data),
	)
	// Load json config file
	tmpConfig.Load(
		memorySource,
		file.NewSource(
			file.WithPath(getConfigFile()),
		),
		env.NewSource(),
		flag.NewSource(),
	)
	var conf Config

	tmpConfig.Scan(&conf)

	return conf
}

func getConfigFile() (configFile string) {
	if value, ok := os.LookupEnv("CERT_OS_CONFIG"); ok {
		return value
	}
	return defaultConfigFile
}

func (c *Config) String() string {
	var s string
	for _, element := range c.Notifiers {
		s += element.name + "\n"
	}
	s += c.Provider.Kind
	return s
}
