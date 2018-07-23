/*
Sample config file

notifiers:
- name: myslacknotifier
  kind: slack
  integration_url: https://hooks.slack.com/services/service_id/auth-token
*/

package stub

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
	"github.com/redhat-cop/cert-operator/pkg/notifier"
)

type Config struct {
	Notifiers map[string]notifier.Notifier `json:"notifiers"`
}

const (
	defaultConfigFile = "/etc/cert-operator/config.yml"
)

func NewConfig() config.Config {
	conf := config.NewConfig()

	// load stuff
	// Load json config file
	config.Load(file.NewSource(
		file.WithPath(defaultConfigFile),
	))
	// done
	return conf
}
