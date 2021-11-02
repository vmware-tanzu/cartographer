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

package registrar_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/cartographer/pkg/registrar"
)

var _ = Describe("Registrar", func() {
	Describe("Now", func() {
		var testTime time.Time
		Context("when Now is called", func() {
			BeforeEach(func() {
				timer := registrar.Timer{}
				testTime = timer.Now().Time
			})

			It("returns the current time", func() {
				Expect(time.Since(testTime)).To(BeNumerically("<", time.Second))
			})
		})
	})

	Describe("AddToScheme", func() {
		Context("when passing in a scheme", func() {
			var scheme *runtime.Scheme
			BeforeEach(func() {
				scheme = runtime.NewScheme()
				Expect(registrar.AddToScheme(scheme)).To(Succeed())
			})

			It("registers the cartographer group", func() {
				Expect(scheme.IsGroupRegistered("carto.run")).To(BeTrue())
			})

			It("creates a scheme with expected length", func() {
				gv := schema.GroupVersion{
					Group:   "carto.run",
					Version: "v1alpha1",
				}
				Expect(len(scheme.KnownTypes(gv))).To(Equal(29))
				// If this test fails, it may indicate that new types should be added to the test below
			})

			It("adds the cartographer objects to the scheme", func() {
				baseGVK := schema.GroupVersionKind{
					Group:   "carto.run",
					Version: "v1alpha1",
				}

				kinds := []string{
					"ClusterConfigTemplate",
					"ClusterDelivery",
					"ClusterDeploymentTemplate",
					"ClusterImageTemplate",
					"ClusterRunTemplate",
					"ClusterSourceTemplate",
					"ClusterSupplyChain",
					"ClusterTemplate",
					"Deliverable",
					"Runnable",
					"Workload",
				}

				for _, kind := range kinds {
					baseGVK.Kind = kind
					Expect(scheme.Recognizes(baseGVK)).To(BeTrue(), fmt.Sprintf("scheme should have kind: %s", kind))
				}
			})
		})
	})
})
