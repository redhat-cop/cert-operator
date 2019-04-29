package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestCertCtrl(t *testing.T) {
	routeList := &routev1.RouteList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
	}
	err := framework.AddToFrameworkScheme(routev1.AddToScheme, routeList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("test-group", func(t *testing.T) {
		t.Run("Cluster", SetupCluster)
		t.Run("Cluster2", SetupCluster)
	})
}

func routeBasicTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleRoute := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "route-tls",
			Namespace: namespace,
			Annotations: map[string]string{
				"openshift.io/cert-ctl-status": "new",
			},
		},
		Spec: routev1.RouteSpec{
			Host: fmt.Sprintf("route-tls.%s.example.com", namespace),
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
			},
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "myservice",
			},
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleRoute, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	assert.Nil(t, waitForAnnotation(t, f, exampleRoute, "openshift.io/cert-ctl-status", "secured"))

	return nil
}

func serviceP12Test(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleService := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-service-pkcs12",
			Namespace: namespace,
			Annotations: map[string]string{
				"openshift.io/cert-ctl-status": "new",
				"openshift.io/cert-ctl-format": "PKCS12",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "web",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"name": "example-service-pkcs12",
			},
			Type: "ClusterIP",
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleService, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	// Get the service to confirm it was created
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: exampleService.Name, Namespace: namespace}, exampleService)
	if err != nil {
		return err
	}

	// Check the the annotation was set properly
	assert.Nil(t, waitForAnnotation(t, f, exampleService, "openshift.io/cert-ctl-status", "secured"))

	// Verify that a secret was created
	exampleSecret := &corev1.Secret{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: exampleService.Name + "-certificate", Namespace: namespace}, exampleSecret)
	if err != nil {
		return err
	}

	// Check that the secret is of the expected type and has values set
	assert.Equal(t, exampleSecret.Type, corev1.SecretTypeOpaque, "invalid secret type")
	assert.NotContains(t, exampleSecret.Data, "tls.crt", "should not have contained tls.crt")
	assert.NotContains(t, exampleSecret.Data, "tls.key", "should not have contained tls.key")
	assert.Contains(t, exampleSecret.Data, "tls.p12", "should have contained tls.p12")
	assert.Contains(t, exampleSecret.Data, "tls-p12-secret.txt", "should have contained tls-p12-secret.txt")

	return nil
}

func serviceBasicTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	exampleService := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-service",
			Namespace: namespace,
			Annotations: map[string]string{
				"openshift.io/cert-ctl-status": "new",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:       "web",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"name": "example-service",
			},
			Type: "ClusterIP",
		},
	}

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), exampleService, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	// Get the service to confirm it was created
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: exampleService.Name, Namespace: namespace}, exampleService)
	if err != nil {
		return err
	}

	// Check the the annotation was set properly
	assert.Nil(t, waitForAnnotation(t, f, exampleService, "openshift.io/cert-ctl-status", "secured"))

	// Verify that a secret was created
	exampleSecret := &corev1.Secret{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: exampleService.Name + "-certificate", Namespace: namespace}, exampleSecret)
	if err != nil {
		return err
	}

	// Check that the secret is of the expected type and has values set
	assert.Equal(t, exampleSecret.Type, corev1.SecretTypeTLS, "invalid secret type")
	assert.Contains(t, exampleSecret.Data, "tls.crt", "did not contain tls.crt")
	assert.Contains(t, exampleSecret.Data, "tls.key", "did not contain tls.key")
	assert.NotContains(t, exampleSecret.Data, "tls.p12", "should not have contained tls.p12")
	assert.NotContains(t, exampleSecret.Data, "tls-p12-secret.txt", "should not have contained tls-p12-secret.txt")

	return nil
}

func SetupCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := framework.Global

	// wait for operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "cert-operator", 3, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = routeBasicTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
	if err = serviceBasicTest(t, f, ctx); err != nil {
		t.Fatal("PEM Service", err)
	}
	if err = serviceP12Test(t, f, ctx); err != nil {
		t.Fatal("PKCS12 Service", err)
	}
}

func waitForAnnotation(t *testing.T, f *framework.Framework, obj IdentifiableKind, annotation string, expectedValue string) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {

		var instance IdentifiableKind
		switch obj.(type) {
		case *routev1.Route:
			instance = &routev1.Route{}
		case *corev1.Service:
			instance = &corev1.Service{}
		case *corev1.Secret:
			instance = &corev1.Secret{}
		default:
			t.Logf("Unsupported kind: %s", obj.GetObjectKind())
			return false, nil
		}

		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, instance.(runtime.Object))
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s deployment\n", obj.GetName())
				return false, nil
			}
			return false, err
		}

		if instance.GetAnnotations()[annotation] == expectedValue {
			return true, nil
		}
		t.Logf("Waiting for operator to reconcile Route %s (current %s; want %s)\n", instance.GetName(), instance.GetAnnotations()[annotation], expectedValue)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Object reconciled! %s\n", obj.GetName())
	return nil
}
