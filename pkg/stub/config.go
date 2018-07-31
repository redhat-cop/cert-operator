/*
Sample config file

notifiers:
- name: myslacknotifier
  kind: slack
  integration_url: https://hooks.slack.com/services/service_id/auth-token
*/

package stub

import (
	"fmt"
	"os"

	config "github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"github.com/micro/go-config/source/flag"
	"github.com/micro/go-config/source/memory"
	"github.com/redhat-cop/cert-operator/pkg/certs"
	"github.com/sirupsen/logrus"
)

type Config struct {
	//	Notifiers []notifier.Notifier  `json:"notifiers"`
	Provider certs.ProviderConfig `json:"provider"`
	General  GeneralConfig        `json:"general"`
}

type GeneralConfig struct {
	Annotations AnnotationConfig `json:"annotations"`
}

type AnnotationConfig struct {
	Status       string `json:"status"`
	StatusReason string `json:"status-reason"`
	Expiry       string `json:"expiry"`
	Format       string `json:"format"`
}

const (
	defaultConfigFile = "/etc/cert-operator/config.yml"
	defaultProvider   = "self-signed"
	defaultConfig     = `
  {
    "general": {
      "annotations": {
        "status": "openshift.io/cert-ctl-status",
        "status-reason": "openshift.io/cert-ctl-status-reason",
        "expiry": "openshift.io/cert-ctl-expires",
        "format": "openshift.io/cert-ctl-format"
      }
    },
    "provider": {
      "kind": "self-signed"
    },
    "notifiers": [
      {
        "name": "log-notifier",
        "log_prefix": "prefix"
      },
      {
        "name": "other",
        "log_prefix": "other-prefix"
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

	fmt.Println("Config: ", tmpConfig.Map()["notifiers"])

	// if conf.Notifiers == nil {
	// 	panic("Notifiers should not be empty")
	// }
	//
	// for index, n := range conf.Notifiers {
	// 	logrus.Infof("Found notifier: " + string(index) + "=" + n.Name)
	// }

	return conf
}

func getConfigFile() (configFile string) {
	if value, ok := os.LookupEnv("CERT_OS_CONFIG"); ok {
		logrus.Infof("Loading config file from %v", value)
		return value
	}
	logrus.Infof("Loading config file from %v", defaultConfigFile)
	return defaultConfigFile
}

func (c *Config) String() string {
	var s string
	// for _, element := range c.Notifiers {
	// 	s += element.Name() + "\n"
	// }
	s += c.Provider.Kind
	return s
}
