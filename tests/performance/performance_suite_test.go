package performance_test

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPerformance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Suite")
}

var (
	c                client.Client
)

var _ = BeforeSuite(func() {
	var err error

	// --- create client
	scheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = resources.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	apiConfig, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(apiConfig, client.Options{
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred())
})
