package stub

import (
	"context"
	"encoding/json"
	"time"

	v1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/redhat-cop/cert-operator/pkg/certs"
	"github.com/redhat-cop/cert-operator/pkg/notifier"
	"github.com/sirupsen/logrus"
)

func NewHandler(config Config) sdk.Handler {
	var provider certs.Provider
	switch config.Provider.Kind {
	case "self-signed":
		provider = new(certs.SelfSignedProvider)
	default:
		panic("No cert provider configured." + config.String())
	}
	return &Handler{
		config:   config,
		provider: provider,
	}
}

type Handler struct {
	// Fill me
	config   Config
	provider certs.Provider
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.Route:
		route := o
		if route.ObjectMeta.Annotations == nil || route.ObjectMeta.Annotations["openshift.io/managed.cert"] == "" {
			return nil
		}

		if route.ObjectMeta.Annotations["openshift.io/managed.cert"] == "new" || route.ObjectMeta.Annotations["openshift.io/managed.cert"] == "renew" {
			// Notfiy of certificate awaiting creation
			logrus.Infof("Found a route waiting for a cert : %v/%v",
				route.ObjectMeta.Namespace,
				route.ObjectMeta.Name)
			notify(route)
			// Create cert request
			oneYear, timeErr := time.ParseDuration("8760h")
			if timeErr != nil {
				return timeErr
			}

			cert, key, err := h.provider.Provision(route.Spec.Host, time.Now().Format("Jan 2 15:04:05 2006"), oneYear, false, 2048, "")
			if err != nil {
				return err
			}

			// Retreive cert from provider
			var routeCopy *v1.Route
			routeCopy = route.DeepCopy()
			routeCopy.ObjectMeta.Annotations["openshift.io/managed.cert"] = "no"
			routeCopy.Spec.TLS = &v1.TLSConfig{
				Termination: v1.TLSTerminationEdge,
				Certificate: string(cert),
				Key:         string(key),
			}
			updateRoute(routeCopy)

			logrus.Infof("Update route %v/%v with new certificate",
				route.ObjectMeta.Namespace,
				route.ObjectMeta.Name)
		}

	}
	return nil
}

func notify(route *v1.Route) {
	result, err := notifier.Notify(route)
	if err != nil {
		panic(err)
	}
	var rm notifier.ResultMessage
	json.Unmarshal(result, &rm)
	if rm.ErrorLog != "" {
		logrus.Errorf(rm.ErrorLog)
	}
	if rm.InfoLog != "" {
		logrus.Infof(rm.InfoLog)
	}
	if rm.DebugLog != "" {
		logrus.Debugf(rm.DebugLog)
	}

}

// update route def
func updateRoute(route *v1.Route) error {

	err := sdk.Update(route)

	return err
}
