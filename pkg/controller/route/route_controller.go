package route

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	v1 "github.com/openshift/api/route/v1"
	"github.com/redhat-cop/cert-operator/pkg/certs"
	certconf "github.com/redhat-cop/cert-operator/pkg/config"
	"github.com/redhat-cop/cert-operator/pkg/helpers"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_route")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Route Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, config certconf.Config) error {
	return add(mgr, newReconciler(mgr, config))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, config certconf.Config) reconcile.Reconciler {
	var provider certs.Provider

	if config.Provider.Ssl == "true" {
		// logrus.Infof("SSL Verified")
		log.Info("SSL Verified")
	} else {
		// logrus.Infof("SSL Not Verified")
		log.Info("SSL Not Verified")
	}

	switch config.Provider.Kind {
	case "none":
		// logrus.Infof("None provider.")
		log.Info("None provider.")
		provider = new(certs.NoneProvider)
	case "self-signed":
		// logrus.Infof("Self Signed provider.")
		log.Info("Self Signed provider.")
		provider = new(certs.SelfSignedProvider)
	case "venafi":
		// logrus.Infof("Venafi Cert provider.")
		provider = new(certs.VenafiProvider)
	default:
		panic("There was a problem detecting which provider to configure. \n" +
			"\tProvider kind `" + config.Provider.Kind + "` is invalid. \n" +
			config.String())
	}

	return &ReconcileRoute{client: mgr.GetClient(), scheme: mgr.GetScheme(), config: config, provider: provider}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("route-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Route
	err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Route
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &routev1.Route{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRoute{}

// ReconcileRoute reconciles a Route object
type ReconcileRoute struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	config   certconf.Config
	provider certs.Provider
}

// Reconcile reads that state of the cluster for a Route object and makes changes based on the state read
// and what is in the Route.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRoute) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the Route route
	route := &routev1.Route{}
	err := r.client.Get(context.TODO(), request.NamespacedName, route)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if route.ObjectMeta.Annotations == nil || route.ObjectMeta.Annotations[r.config.General.Annotations.Status] == "" {
		return reconcile.Result{}, nil
	}

	if route.ObjectMeta.Annotations[r.config.General.Annotations.Status] == r.config.General.Annotations.NeedCertValue {
		reqLogger.Info("Reconciling Route")

		if route.Spec.TLS.Termination == v1.TLSTerminationPassthrough {
			route.ObjectMeta.Annotations[r.config.General.Annotations.Status] = "failed"
			route.ObjectMeta.Annotations[r.config.General.Annotations.StatusReason] = "Certificate and key cannot be set on Passthrough route"

			err = helpers.Apply(r.client, route)
			return reconcile.Result{}, err
		}

		// Retrieve cert from provider
		keyPair, err := helpers.GetCert(route.Spec.Host, r.provider, r.config.Provider.Ssl)
		if err != nil {
			route.ObjectMeta.Annotations[r.config.General.Annotations.Status] = "failed"
			route.ObjectMeta.Annotations[r.config.General.Annotations.StatusReason] = err.Error()
		} else {
			route.ObjectMeta.Annotations[r.config.General.Annotations.Status] = "secured"
			route.ObjectMeta.Annotations[r.config.General.Annotations.Expiry] = keyPair.Expiry.Format(helpers.TimeFormat)
		}

		//var termination string

		var termination v1.TLSTerminationType
		config := route.Spec.TLS
		if config == nil {
			termination = v1.TLSTerminationEdge
		} else {
			termination = route.Spec.TLS.Termination
		}

		route.Spec.TLS = &v1.TLSConfig{
			Termination: termination,
			Certificate: string(keyPair.Cert),
			Key:         string(keyPair.Key),
		}

		err = helpers.Apply(r.client, route)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Updated route with new certificate")
	}

	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *routev1.Route) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
