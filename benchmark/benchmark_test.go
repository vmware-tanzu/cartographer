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
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/vmware-tanzu/cartographer/benchmark/sampler"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
	"github.com/vmware-tanzu/cartographer/tests/resources"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
	"strconv"
	"time"
)

const timeoutSecsPerWorkload = 10
const pollingFrequencySecs = 5
const numPollingIntervalsToWaitAtEnd = 5

var createdObjects []*unstructured.Unstructured

func makeSupplyChain(name string) {
	yamlString := utils.HereYaml(`
		---
		apiVersion: carto.run/v1alpha1
		kind: ClusterTemplate
		metadata:
		  name: "` + name + `-template"
		spec:
		  template:
			apiVersion: v1
			kind: ConfigMap
			metadata:
			  name: $(workload.metadata.name)$-stamped-object
			  labels: {} #$(workload.metadata.labels)$
			data:
			  foo: bar
	`)
	template := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(yamlString), template)).To(Succeed())
	Expect(c.Create(context.TODO(), template)).To(Succeed())
	createdObjects = append(createdObjects, template)

	yamlString = utils.HereYaml(`---
		apiVersion: carto.run/v1alpha1
		kind: ClusterSupplyChain
		metadata:
		  name: "` + name + `"
		spec:
		  selector:
			for-supply-chain: "` + name + `"
		  resources:
			- name: template
			  templateRef:
				name: "` + name + `-template"
				kind: ClusterTemplate
	`)
	supplychain := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(yamlString), supplychain)).To(Succeed())
	Expect(c.Create(context.TODO(), supplychain)).To(Succeed())
	createdObjects = append(createdObjects, supplychain)
}

func makeWorkload(name, supplychain string) {
	yamlString := utils.HereYaml(`---
        apiVersion: carto.run/v1alpha1
        kind: Workload
        metadata:
          name: "` + name + `"
          namespace: default
          labels:
            for-supply-chain: "` + supplychain + `"
        spec: {}
	`)
	workload := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(yamlString), workload)).To(Succeed())
	Expect(c.Create(context.TODO(), workload)).To(Succeed())
	createdObjects = append(createdObjects, workload)
}

func makeRunnable(name string) {
	yamlString := utils.HereYaml(`---
        apiVersion: carto.run/v1alpha1
        kind: ClusterRunTemplate
		metadata:
          name: "` + name + `-runnable-template"
		spec:
          outputs:
            my-output: .spec.foo
		  template:
		    apiVersion: test.run/v1alpha1
		    kind: TestObj
		    metadata:
		      generateName: "` + name + `-stamped-object"
            spec:
              foo: "value is $(runnable.spec.inputs.arg)$"
		    status:
		      conditions:
		      - type: Succeeded
		        status: "Unknown"
	`)
	clusterruntemplate := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(yamlString), clusterruntemplate)).To(Succeed())
	Expect(c.Create(context.TODO(), clusterruntemplate)).To(Succeed())
	createdObjects = append(createdObjects, clusterruntemplate)

	yamlString = utils.HereYaml(`---
        apiVersion: carto.run/v1alpha1
        kind: Runnable
		metadata:
		  name: "` + name + `"
          namespace: default
		spec:	
		  runTemplateRef:
		    name: "` + name + `-runnable-template"
		  inputs:
            command: "echo"
            arg: "0"
	`)
	runnable := &unstructured.Unstructured{}
	Expect(yaml.Unmarshal([]byte(yamlString), runnable)).To(Succeed())
	Expect(c.Create(context.TODO(), runnable)).To(Succeed())
	createdObjects = append(createdObjects, runnable)
}

func GetCartographerControllerPod() types.NamespacedName {
	allPods := &v1.PodList{}
	Expect(c.List(context.TODO(), allPods, client.MatchingLabels(map[string]string{"app": "cartographer-controller"}))).To(Succeed())
	Expect(len(allPods.Items)).To(Equal(1))
	return types.NamespacedName{
		Namespace: allPods.Items[0].GetNamespace(),
		Name:      allPods.Items[0].GetName(),
	}
}

var _ = Describe("Benchmark", func() {
	var (
		numWorkloads    int
		numSupplyChains int
		numRunnables    int
		numRunnableRuns int
		err             error
		s               sampler.PodMetricsSampler
	)

	BeforeEach(func() {
		numRunnables = 10
		numSupplyChains = 3
		numWorkloads = 10
		numRunnableRuns = 10

		if nSupplyChainsStr, ok := os.LookupEnv("CARTO_BENCH_N_SUPPLY_CHAINS"); ok {
			numSupplyChains, err = strconv.Atoi(nSupplyChainsStr)
			Expect(err).NotTo(HaveOccurred())
		}
		if nWorkloadsStr, ok := os.LookupEnv("CARTO_BENCH_N_WORKLOADS"); ok {
			numWorkloads, err = strconv.Atoi(nWorkloadsStr)
			Expect(err).NotTo(HaveOccurred())
		}
		if nRunnablesStr, ok := os.LookupEnv("CARTO_BENCH_N_RUNNABLES"); ok {
			numRunnables, err = strconv.Atoi(nRunnablesStr)
			Expect(err).NotTo(HaveOccurred())
		}
		if nRunnableRunsStr, ok := os.LookupEnv("CARTO_BENCH_N_RUNNABLE_RUNS"); ok {
			numRunnableRuns, err = strconv.Atoi(nRunnableRunsStr)
			Expect(err).NotTo(HaveOccurred())
		}

		fmt.Printf("--- Starting benchmark, #supply chains:%d #workloads:%d #runnables:%d #runs-per-runnable:%d ---\n", numSupplyChains, numWorkloads, numRunnables, numRunnableRuns)

		s, err = sampler.NewPodMetricsSampler(c, fmt.Sprintf("%dscs_%dworkloads_%drunnables_%druns", numSupplyChains, numWorkloads, numRunnables, numRunnableRuns), GetCartographerControllerPod())
		Expect(err).NotTo(HaveOccurred())

		testObjCrdBytes, err := os.ReadFile("../tests/resources/crds/test.run_testobjs.yaml")
		Expect(err).NotTo(HaveOccurred())

		testObjCrd := &unstructured.Unstructured{}
		Expect(yaml.Unmarshal(testObjCrdBytes, testObjCrd)).To(Succeed())
		Expect(c.Create(context.TODO(), testObjCrd)).To(Succeed())
		createdObjects = append(createdObjects, testObjCrd)
	})

	AfterEach(func() {
		Expect(s.SampleConsumption("Before cleanup")).To(Succeed())
		for _, obj := range createdObjects {
			_ = c.Delete(context.TODO(), obj)
		}
		Expect(s.SampleConsumption("After cleanup"))

		fmt.Printf("Memory: %v min, %v max\nCPU: %v min, %v max", s.MinMem(), s.MaxMem(), s.MinCPU(), s.MaxCPU())
	})

	Context("Supplychain", func() {
		BeforeEach(func() {
			Expect(s.SampleConsumption(fmt.Sprintf("Before creating %d supply chains", numSupplyChains))).To(Succeed())
			for i := 0; i < numSupplyChains; i++ {
				makeSupplyChain(fmt.Sprintf("bar-supplychain-%d", i))
			}

			Expect(s.SampleConsumption(fmt.Sprintf("Before creating %d workloads", numWorkloads))).To(Succeed())
			for i := 0; i < numWorkloads; i++ {
				makeWorkload(fmt.Sprintf("foo-workload-%d", i), fmt.Sprintf("bar-supplychain-%d", i%numSupplyChains))
			}

			Expect(s.SampleConsumption(fmt.Sprintf("Before creating %d runnables", numRunnables))).To(Succeed())
			for i := 0; i < numRunnables; i++ {
				makeRunnable(fmt.Sprintf("foo-runnable-%d", i))
			}

			Expect(s.SampleConsumption("Creation of supply chains, workloads and runnables complete")).To(Succeed())
		})

		It("eventually completes processing all workloads and runnables", func() {
			completedWorkloads := map[int]bool{}
			completedRunnables := map[int]bool{}
			completedRunnableRuns := map[int]int{}

			Eventually(func(g Gomega) {
				//Sample consumption
				Expect(s.SampleConsumption(fmt.Sprintf("Polling. Previously completed: [%d/%d workloads] [%d/%d runnables]", len(completedWorkloads), numWorkloads, len(completedRunnables), numRunnables))).To(Succeed())

				//Check for finished runnables
				for nRunnable := 0; nRunnable < numRunnables; nRunnable++ {
					if completedRunnableRuns[nRunnable] == numRunnableRuns {
						completedRunnables[nRunnable] = true
						continue
					}
					runnable := &v1alpha1.Runnable{}
					key := client.ObjectKey{Name: fmt.Sprintf("foo-runnable-%d", nRunnable), Namespace: "default"}
					g.Expect(c.Get(context.Background(), key, runnable)).To(Succeed())
					g.Expect(runnable).ToNot(BeNil())

					succeeded := false
					for _, condition := range runnable.Status.Conditions {
						if condition.Type == "Ready" && condition.Status == "True" {
							expectedOutput := fmt.Sprintf(`"value is %d"`, completedRunnableRuns[nRunnable])
							actualOutput := string(runnable.Status.Outputs["my-output"].Raw)
							if runnable.Status.Outputs != nil && actualOutput == expectedOutput {
								succeeded = true
								completedRunnableRuns[nRunnable]++
								if completedRunnableRuns[nRunnable] < numRunnableRuns {
									nextRunnableRunNumber := completedRunnableRuns[nRunnable] + 1
									Expect(s.SampleConsumption(fmt.Sprintf("In runnable %d's poll, triggering run #%d", nRunnable, nextRunnableRunNumber))).To(Succeed())
									runnable.Spec.Inputs["arg"] = apiextensionsv1.JSON{
										Raw: []byte(fmt.Sprintf("%d", completedRunnableRuns[nRunnable])),
									}
									g.Expect(c.Update(context.Background(), runnable)).To(Succeed())
									succeeded = false
								}
							}
						}
					}

					//if the runnable hasn't succeeded yet, check for a templated object that was stamped without success yet, and simulate a transition to success if it hasn't already happened
					if !succeeded {
						stampedRunnableList := &resources.TestObjList{}
						err = c.List(context.Background(), stampedRunnableList, client.InNamespace("default"), client.MatchingLabels(map[string]string{"carto.run/runnable-name": runnable.Name}))
						g.Expect(err).NotTo(HaveOccurred())

						for _, stampedRunnable := range stampedRunnableList.Items {
							if len(stampedRunnable.Status.Conditions) == 0 {
								Expect(s.SampleConsumption(fmt.Sprintf("In runnable %d's poll, simulating successful transition for run #%d", nRunnable, len(stampedRunnableList.Items)))).To(Succeed())

								stampedRunnable.Status.Conditions = []metav1.Condition{
									{
										Type:               "Succeeded",
										Status:             "True",
										ObservedGeneration: stampedRunnable.Generation,
										LastTransitionTime: metav1.Time{Time: time.Now()},
										Reason:             "because",
										Message:            "see?",
									},
								}
								stampedRunnable.Status.ObservedGeneration = stampedRunnable.Generation
								g.Expect(c.Status().Update(context.Background(), &stampedRunnable)).To(Succeed())
							}
						}
						continue
					}

					g.Expect(completedRunnableRuns[nRunnable]).To(Equal(numRunnableRuns))
					Expect(s.SampleConsumption(fmt.Sprintf("Runnable %d succeeded with outputs from %d runs\n", nRunnable, numRunnableRuns))).To(Succeed())
				}
				g.Expect(len(completedRunnables)).To(Equal(numRunnables))

				//Check for finished workloads
				for nWorkload := 0; nWorkload < numWorkloads; nWorkload++ {
					if completedWorkloads[nWorkload] {
						continue
					}
					workload := &v1alpha1.Workload{}
					key := client.ObjectKey{Name: fmt.Sprintf("foo-workload-%d", nWorkload), Namespace: "default"}
					g.Expect(c.Get(context.Background(), key, workload)).To(Succeed())

					succeeded := false
					for _, condition := range workload.Status.Conditions {
						if condition.Type == "Ready" && condition.Status == "True" {
							succeeded = true
						}
					}
					g.Expect(succeeded).To(BeTrue())

					//Confidence check the templated object was stamped.
					stamped := &v1.ConfigMap{}
					key = client.ObjectKey{Name: fmt.Sprintf("foo-workload-%d-stamped-object", nWorkload), Namespace: "default"}
					err = c.Get(context.Background(), key, stamped)
					g.Expect(err).NotTo(HaveOccurred())

					Expect(s.SampleConsumption(fmt.Sprintf("workload %d finished", nWorkload))).To(Succeed())

					completedWorkloads[nWorkload] = true
				}
			}).WithTimeout(time.Duration(timeoutSecsPerWorkload*(numWorkloads+(numRunnables*numRunnableRuns))) * time.Second).WithPolling(pollingFrequencySecs * time.Second).Should(Succeed())

			By("Observing stability after all initial workloads and runnables are reconciled...")
			for i := 0; i < numPollingIntervalsToWaitAtEnd; i++ {
				time.Sleep(pollingFrequencySecs * time.Second)
				Expect(s.SampleConsumption(fmt.Sprintf("Waiting %d/%d", i, numPollingIntervalsToWaitAtEnd))).To(Succeed())
			}
		})
	})
})
