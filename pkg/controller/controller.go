package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"

	certconf "github.com/redhat-cop/cert-operator/pkg/config"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, certconf.Config) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, c certconf.Config) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, c); err != nil {
			return err
		}
	}
	return nil
}
