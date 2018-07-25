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
)

type OpConfig struct {
	Notifiers map[string]string `json:"notifiers"`
}

const (
	defaultConfigFile = "/etc/cert-operator/config.yml"
)

// func NewConfig() OpConfig {
// 	// Load json config file
// 	config.Load(
// 		env.NewSource(),
// 		flag.NewSource(),
// 		file.NewSource(
// 			file.WithPath(getConfigFile()),
// 		),
// 	)
//
// 	var notifiers config.Config
//
// 	//	config.get
//
// 	return &OpConfig{
// 		Notifiers: config.Map()["notifiers"],
// 	}
// }

func getConfigFile() (configFile string) {
	if value, ok := os.LookupEnv("CERT_OS_CONFIG"); ok {
		return value
	}
	return defaultConfigFile
}
