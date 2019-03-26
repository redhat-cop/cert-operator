package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// get global framework variables
	f := framework.Global

	if err = routeBasicTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
