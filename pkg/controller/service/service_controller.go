package service

import (
	"context"
	"encoding/pem"

	"github.com/redhat-cop/cert-operator/pkg/certs"
	certconf "github.com/redhat-cop/cert-operator/pkg/config"
	"github.com/redhat-cop/cert-operator/pkg/helpers"
	"github.com/redhat-cop/cert-operator/pkg/rand"
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

var log = logf.Log.WithName("controller_service")

// Add creates a new Service Controller and adds it to the Manager. The Manager will set fields on the Controller
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

	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme(), config: config, provider: provider}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Route
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.Service{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileRoute reconciles a Route object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	config   certconf.Config
	provider certs.Provider
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Route.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the Route svc
	svc := &corev1.Service{}
	err := r.client.Get(context.TODO(), request.NamespacedName, svc)
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

	// Look for annoation that requires action, otherwise skip it
	if svc.ObjectMeta.Annotations == nil || svc.ObjectMeta.Annotations[r.config.General.Annotations.Status] == "" {
		return reconcile.Result{}, nil
	}

	if svc.ObjectMeta.Annotations[r.config.General.Annotations.Status] == r.config.General.Annotations.NeedCertValue {
		reqLogger.Info("Reconciling Service")

		host := svc.ObjectMeta.Name + "." + svc.ObjectMeta.Namespace + ".svc"

		keyPair, err := helpers.GetCert(host, r.provider, r.config.Provider.Ssl)
		if err != nil {
			svc.ObjectMeta.Annotations[r.config.General.Annotations.Status] = "failed"
			svc.ObjectMeta.Annotations[r.config.General.Annotations.StatusReason] = err.Error()
		} else {
			svc.ObjectMeta.Annotations[r.config.General.Annotations.Status] = "secured"
			svc.ObjectMeta.Annotations[r.config.General.Annotations.Expiry] = keyPair.Expiry.Format(helpers.TimeFormat)
		}

		dm := make(map[string][]byte)
		secretType := corev1.SecretTypeTLS

		// see what format was requested
		if svc.ObjectMeta.Annotations[r.config.General.Annotations.Format] == r.config.General.Annotations.Pkcs12Format {
			password := rand.String(24)
			pemCrt, _ := pem.Decode(keyPair.Cert)
			pemKey, _ := pem.Decode(keyPair.Key)
			p12cert, err := certs.ConvertToPKCS12(pemKey.Bytes, pemCrt.Bytes, [][]byte{}, password)

			if err != nil {
				reqLogger.Error(err, "Failed to convert to PKCS12")
				return reconcile.Result{}, err
			}

			dm["tls.p12"] = p12cert
			dm["tls-p12-secret.txt"] = []byte(password)

			// override secret type since it's not tls
			secretType = corev1.SecretTypeOpaque
		} else {
			dm["tls.crt"] = keyPair.Cert
			dm["tls.key"] = keyPair.Key
		}

		// Create a secret
		certSec := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.ObjectMeta.Name + "-certificate",
				Namespace: svc.ObjectMeta.Namespace,
			},
			Data: dm,
			Type: secretType,
		}

		err = helpers.Apply(r.client, certSec)
		if err != nil {
			reqLogger.Error(err, "Failed to apply secret")
			return reconcile.Result{}, err
		}

		err = helpers.Apply(r.client, svc)
		if err != nil {
			reqLogger.Error(err, "Failed to apply service")
			return reconcile.Result{}, err
		}

		reqLogger.Info("Updated service with new certificate")
	}

	return reconcile.Result{}, nil
}
