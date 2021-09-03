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

package supply_chain_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/root"
)

func TestWebhookIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
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
		testEnv.ControlPlane.APIServer.Configure().
			Append("audit-policy-file", filepath.Join(workingDir, "policy.yaml")).
			Append("audit-log-path", "-")
	}

	apiConfig, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	// get a kubeconfig
	kubeConfigFile, err := generateConfigFile(testEnv)
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

	Eventually(controllerBuffer, 10*time.Second).Should(gbytes.Say("serving webhook server"))
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("Server: %s", testEnv.WebhookInstallOptions.LocalServingHost)

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
	err := ensureNamespace(testNS, c)
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

func generateConfigFile(env *envtest.Environment) (string, error) {
	user, err := env.ControlPlane.AddUser(envtest.User{
		Name:   "envtest-admin",
		Groups: []string{"system:masters"},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("add user: %w", err)
	}

	kubeconfigFile, err := ioutil.TempFile("", "cartographer-integration-test-kubeconfig-")
	if err != nil {
		return "", fmt.Errorf("tempfile: %w", err)
	}

	kubeConfig, err := user.KubeConfig()
	if err != nil {
		return "", fmt.Errorf("kubeconfig: %w", err)
	}

	if _, err := kubeconfigFile.Write(kubeConfig); err != nil {
		return "", fmt.Errorf("write kubeconfig: %w", err)
	}

	return kubeconfigFile.Name(), nil
}

func ensureNamespace(namespace string, client client.Client) error {
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
	}
	err := client.Create(context.TODO(), &ns)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
