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

package benchmark_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" //FIXME:seriously, for GCP to work? nb: plugins for azure/aws probably needed on those iaas
	metrics "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"
)

func TestPerformance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Suite")
}

var (
	c client.Client
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

	Expect(metrics.AddToScheme(scheme)).To(Succeed())

	apiConfig, err := config.GetConfig()
	Expect(err).NotTo(HaveOccurred())

	c, err = client.New(apiConfig, client.Options{
		Scheme: scheme,
	})
	Expect(err).NotTo(HaveOccurred())
})
