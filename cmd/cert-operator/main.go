package main

import (
	"context"
	"runtime"

	route "github.com/openshift/api/route/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	stub "github.com/redhat-cop/cert-operator/pkg/stub"

	"github.com/sirupsen/logrus"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()
	k8sutil.AddToSDKScheme(route.AddToScheme)

	conf := stub.NewConfig()
	sdk.ExposeMetricsPort()

	logrus.Infof("Watching Routes on all Namespaces")
	sdk.Watch("route.openshift.io/v1", "Route", "", 1000000000)
	logrus.Infof("Watching Services on all Namespaces")
	sdk.Watch("v1", "Service", "", 1000000000)
	sdk.Handle(stub.NewHandler(conf))
	sdk.Run(context.TODO())
}
