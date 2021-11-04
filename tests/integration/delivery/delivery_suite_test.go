// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package delivery_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/root"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
)

func TestDeliveryIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Delivery Integration Suite")
}

var (
	testEnv          *envtest.Environment
	c                client.Client
	testNS           string
	workingDir       string
	cancel           context.CancelFunc
	controllerError  chan error
	controller       *root.Command
	controllerBuffer *gbytes.Buffer
)

const DebugControlPlane = false

var _ = BeforeSuite(func() {
	var err error
	workingDir, err = os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	// start kube-apiserver and etcd
	testEnv = &envtest.Environment{
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "config", "webhook")},
		},
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		AttachControlPlaneOutput: DebugControlPlane, // Set to true for great debug logging
	}

	if DebugControlPlane {
		testEnv.ControlPlane.GetAPIServer().Configure().
			Append("audit-policy-file", filepath.Join(workingDir, "policy.yaml")).
			Append("audit-log-path", "-")
	}

	apiConfig, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	// get a kubeconfig
	kubeConfigFile, err := helpers.GenerateConfigFile(testEnv)
	Expect(err).NotTo(HaveOccurred())

	err = os.Setenv("KUBECONFIG", kubeConfigFile)
	Expect(err).NotTo(HaveOccurred())

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	controllerBuffer = gbytes.NewBuffer()
	controllerOutput := io.MultiWriter(controllerBuffer, GinkgoWriter)
	logger := zap.New(zap.WriteTo(controllerOutput))

	controllerError = make(chan error)

	go func() {
		controller = &root.Command{
			Port:    testEnv.WebhookInstallOptions.LocalServingPort,
			CertDir: testEnv.WebhookInstallOptions.LocalServingCertDir,
			Context: ctx,
			Logger:  logger,
		}

		controllerError <- controller.Execute()
	}()

	// Can take a long time to start serving
	// FIXME: use a real health check, not log line detection

	Eventually(controllerBuffer, 10*time.Second).Should(gbytes.Say("Starting Controller"))
	time.Sleep(200 * time.Millisecond)

	// --- create client
	scheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = batchv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(apiConfig, client.Options{
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred())
})

var _ = BeforeEach(func() {
	testNS = names.SimpleNameGenerator.GenerateName("testns-")
	err := helpers.EnsureNamespace(testNS, c)
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: testNS,
		},
	}
	err := c.Delete(context.Background(), ns, &client.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancel()
	//Expect(<-controllerError).NotTo(HaveOccurred())  // TODO Figure out how to gracefully exit

	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())

	gexec.CleanupBuildArtifacts()
})
