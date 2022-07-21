package controllers_test

import (
	"testing"

	blueprintsv1alpha1 "carto.run/blueprints/api/v1alpha1"
	"carto.run/blueprints/controllers"
	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	rtesting "github.com/vmware-labs/reconciler-runtime/testing"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestInMemoryGatewayReconciler(t *testing.T) {

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = blueprintsv1alpha1.AddToScheme(scheme)

	rts := rtesting.ReconcilerTestSuite{{
		Name: "nothing on cluster",
	}, {
		Name: "valid simple schema",
		Request: ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: "test-ns",
				Name:      "acme.url",
			},
		},
		GivenObjects:            nil,
		APIGivenObjects:         nil,
		ExpectTracks:            nil,
		ExpectEvents:            nil,
		ExpectCreates:           nil,
		ExpectUpdates:           nil,
		ExpectPatches:           nil,
		ExpectDeletes:           nil,
		ExpectDeleteCollections: nil,
		ExpectStatusUpdates:     nil,
		ExpectStatusPatches:     nil,
		AdditionalConfigs:       nil,
		ShouldErr:               false,
		ExpectedResult:          ctrl.Result{},
		Verify:                  nil,
		Prepare:                 nil,
		CleanUp:                 nil,
	}}

	rts.Run(t, scheme, func(t *testing.T, rtc *rtesting.ReconcilerTestCase, c reconcilers.Config) reconcile.Reconciler {
		return controllers.ClusterBlueprintTypeReconciler(c)
	})
}
