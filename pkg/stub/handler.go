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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeFormat = "Jan 2 15:04:05 2006"
)

func NewHandler(config Config) sdk.Handler {
	var provider certs.Provider

	if config.Provider.Ssl == "true" {
		logrus.Infof("SSL Verified")
	} else {
		logrus.Infof("SSL Not Verified")
	}

	switch config.Provider.Kind {
	case "none":
		logrus.Infof("None provider.")
		provider = new(certs.NoneProvider)
	case "self-signed":
		logrus.Infof("Self Signed provider.")
		provider = new(certs.SelfSignedProvider)
	case "venafi":
		logrus.Infof("Venafi Cert provider.")
		provider = new(certs.VenafiProvider)
	default:
		panic("There was a problem detecting which provider to configure. \n" +
			"\tProvider kind `" + config.Provider.Kind + "` is invalid. \n" +
			config.String())
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
		h.handleRoute(o)
	case *corev1.Service:
		h.handleService(o)
	}
	return nil
}

func (h *Handler) handleRoute(route *v1.Route) error {
	if route.ObjectMeta.Annotations == nil || route.ObjectMeta.Annotations[h.config.General.Annotations.Status] == "" {
		return nil
	}

	if route.ObjectMeta.Annotations[h.config.General.Annotations.Status] == h.config.General.Annotations.NeedCertValue {
		// Notfiy of certificate awaiting creation
		logrus.Infof("Found a route waiting for a cert : %v/%v",
			route.ObjectMeta.Namespace,
			route.ObjectMeta.Name)
		message := "" +
			"_Namespace_: *" + route.ObjectMeta.Namespace + "*\n" +
			"_Route Name_: *" + route.ObjectMeta.Name + "*\n"

		h.notify(message)

		// Retreive cert from provider
		keyPair, err := h.getCert(route.Spec.Host)

		var routeCopy *v1.Route
		routeCopy = route.DeepCopy()
		if err != nil {
			routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.Status] = "failed"
			routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.StatusReason] = err.Error()
		} else {
			routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.Status] = "secured"
		}
		routeCopy.ObjectMeta.Annotations[h.config.General.Annotations.Expiry] = keyPair.Expiry.Format(timeFormat)

		config := routeCopy.Spec.TLS
		if config == nil {
			// Create new basic TLS Config
			routeCopy.Spec.TLS = &v1.TLSConfig{
				Termination: v1.TLSTerminationEdge,
				Certificate: string(keyPair.Cert),
				Key:         string(keyPair.Key),
			}
		} else {
			// TLS Config already exists, so we'll just inject a new Cert & Key
			routeCopy.Spec.TLS.Certificate = string(keyPair.Cert)
			routeCopy.Spec.TLS.Key = string(keyPair.Key)
		}

		err = apply(routeCopy)

		if err != nil {
			logrus.Errorf("Error handling route: %s", err.Error())
		}

		logrus.Infof("Updated route %v/%v with new certificate",
			route.ObjectMeta.Namespace,
			route.ObjectMeta.Name)
	}
	return nil
}

func (h *Handler) handleService(service *corev1.Service) error {
	if service.ObjectMeta.Annotations == nil || service.ObjectMeta.Annotations[h.config.General.Annotations.Status] == "" {
		return nil
	}

	if service.ObjectMeta.Annotations[h.config.General.Annotations.Status] == h.config.General.Annotations.NeedCertValue {
		logrus.Infof("Found a service waiting for a cert : %v/%v",
			service.ObjectMeta.Namespace,
			service.ObjectMeta.Name)

		message := "" +
			"_Namespace_: *" + service.ObjectMeta.Namespace + "*\n" +
			"_Service Name_: *" + service.ObjectMeta.Name + "*\n"

		h.notify(message)

		host := service.ObjectMeta.Name + "." + service.ObjectMeta.Namespace + ".svc"

		// Retreive cert from provider
		keyPair, err := h.getCert(host)

		if err != nil {
			logrus.Errorf(err.Error())
		}

		var svcCopy *corev1.Service
		svcCopy = service.DeepCopy()
		svcCopy.ObjectMeta.Annotations[h.config.General.Annotations.Status] = "secured"
		svcCopy.ObjectMeta.Annotations[h.config.General.Annotations.Expiry] = keyPair.Expiry.Format(timeFormat)

		dm := make(map[string][]byte)
		dm["app.crt"] = keyPair.Cert
		dm["app.key"] = keyPair.Key

		// Create a secret
		certSec := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcCopy.ObjectMeta.Name + "-certificate",
				Namespace: svcCopy.ObjectMeta.Namespace,
			},
			Data: dm,
		}

		err = apply(certSec)

		if err != nil {
			logrus.Errorf("Error handling secret: %s", err.Error())
		}

		logrus.Infof("Provisioned new secret %s/%s containing certificate",
			certSec.ObjectMeta.Namespace,
			certSec.ObjectMeta.Name)

		err = apply(svcCopy)

		if err != nil {
			logrus.Errorf("Error handling service: %s", err.Error())
		}

		logrus.Infof("Updated service %v/%v with new certificate",
			service.ObjectMeta.Namespace,
			service.ObjectMeta.Name)
	}		
	return nil
}

func (h *Handler) notify(message string) {
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

func (h *Handler) getCert(host string) (certs.KeyPair, error) {
	oneYear, timeErr := time.ParseDuration("8760h")
	if timeErr != nil {
		logrus.Errorf("Failed to parse time duratio during getCert: " + timeErr.Error())
	}

	// Retreive cert from provider
	keyPair, err := h.provider.Provision(
		host,
		time.Now().Format(timeFormat),
		oneYear, false, 2048, "", h.config.Provider.Ssl)
	if err != nil {
		logrus.Errorf("Failed to provision key pair: " + err.Error())
		return keyPair, err
	}
	return keyPair, nil
}

func apply(object sdk.Object) error {
	err := sdk.Create(object)
	if(err != nil) {
		err = sdk.Update(object)
		if err != nil {
			return err
		}
		return nil;
	}
	return nil
}
