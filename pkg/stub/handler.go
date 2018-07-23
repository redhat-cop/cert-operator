package stub

import (
	"context"
	"encoding/json"

	v1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/redhat-cop/cert-operator/pkg/notifier"
	"github.com/sirupsen/logrus"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
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
			logrus.Infof("Found a route waiting for a cert : %v/%v", route.ObjectMeta.Namespace, route.ObjectMeta.Name)
			notify(route)
			// Create cert request

			// Retreive cert from provider
			routeCopy := route.DeepCopy()
			routeCopy.ObjectMeta.Annotations["openshift.io/managed.cert"] = "no"
			updateRoute(routeCopy)
		}

	}
	return nil
}

func notify(route *v1.Route) {
	url := "https://hooks.slack.com/services/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"
	result, err := notifier.Notify(route, url, "slack")
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
