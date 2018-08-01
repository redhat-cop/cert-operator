package stub

import (
	"context"
	"os"
	"time"

	v1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/redhat-cop/cert-operator/pkg/certs"
	"github.com/redhat-cop/cert-operator/pkg/notifier/slack"
	"github.com/sirupsen/logrus"
)

const (
	timeFormat = "Jan 2 15:04:05 2006"
)

func NewHandler(config Config) sdk.Handler {
	var provider certs.Provider
	switch config.Provider.Kind {
	case "none":
		provider = new(certs.NoneProvider)
	case "self-signed":
		provider = new(certs.SelfSignedProvider)
	default:
		panic("There was a problem detecting which provider to configure. " +
			"Provider kind `" + config.Provider.Kind + "` is invalid.")
	}
	return &Handler{
		config:   config,
		provider: provider,
		//		notifiers: config.Notifiers,
	}
}

type Handler struct {
	// Fill me
	config   Config
	provider certs.Provider
	//	notifiers []notifier.Notifier
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.Route:
		route := o
		if route.ObjectMeta.Annotations == nil || route.ObjectMeta.Annotations[h.config.General.Annotations.Status] == "" {
			return nil
		}

		if route.ObjectMeta.Annotations[h.config.General.Annotations.Status] == "new" {
			// Notfiy of certificate awaiting creation
			logrus.Infof("Found a route waiting for a cert : %v/%v",
				route.ObjectMeta.Namespace,
				route.ObjectMeta.Name)
			h.notify(route)
			// Create cert request
			oneYear, timeErr := time.ParseDuration("8760h")
			if timeErr != nil {
				return timeErr
			}

			keyPair, err := h.provider.Provision(route.Spec.Host, time.Now().Format(timeFormat), oneYear, false, 2048, "")
			if err != nil {
				return err
			}

			// Retreive cert from provider
			var routeCopy *v1.Route
			routeCopy = route.DeepCopy()
			routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.Status] = "no"
			routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.Expiry] = keyPair.Expiry.Format(timeFormat)
			routeCopy.Spec.TLS = &v1.TLSConfig{
				Termination: v1.TLSTerminationEdge,
				Certificate: string(keyPair.Cert),
				Key:         string(keyPair.Key),
			}
			updateRoute(routeCopy)

			logrus.Infof("Updated route %v/%v with new certificate",
				route.ObjectMeta.Namespace,
				route.ObjectMeta.Name)
		}

	}
	return nil
}

func (h *Handler) notify(route *v1.Route) {
	message := "" +
		"_Namespace_: *" + route.ObjectMeta.Namespace + "*\n" +
		"_Route Name_: *" + route.ObjectMeta.Name + "*\n"

	switch os.Getenv("NOTIFIER_TYPE") {
	case "slack":
		c, err := slack.New()
		if err != nil {
			logrus.Errorf("Failed to instantiate notifier\n" + err.Error())
		}
		err = c.Send(message)
		if err != nil {
			logrus.Errorf("Failed to send notification\n" + err.Error())
		} else {
			logrus.Infof("Notification sent: \nMessage:" + message)
		}
	default:
		logrus.Infof("No notification sent, as no notifier is configured.")
	}
}

// update route def
func updateRoute(route *v1.Route) error {

	err := sdk.Update(route)

	return err
}
