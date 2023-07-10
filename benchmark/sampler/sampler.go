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

package sampler

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodMetricsSampler interface {
	Reset()
	SampleConsumption(label string) error
	MaxCPU() *resource.Quantity
	MaxMem() *resource.Quantity
	MinCPU() *resource.Quantity
	MinMem() *resource.Quantity
}

type podMetricsSampler struct {
	consumptionSamples             []PodMetricsEntry
	datasetDescription             string
	podName                        string
	podNamespace                   string
	cl                             client.Client
	maxCpu, maxMem, minMem, minCpu *resource.Quantity
}

type PodMetricsEntry struct {
	Label      string
	PodMetrics *v1beta1.PodMetrics
}

func NewPodMetricsSampler(cl client.Client, datasetDescription string, pod types.NamespacedName) (PodMetricsSampler, error) {
	sampler := &podMetricsSampler{
		consumptionSamples: []PodMetricsEntry{},
		datasetDescription: datasetDescription,
		podName:            pod.Name,
		podNamespace:       pod.Namespace,
		cl:                 cl,
	}
	return sampler, sampler.SampleConsumption(fmt.Sprintf("Start of sample for dataset %s", datasetDescription))
}

func (s *podMetricsSampler) MaxCPU() *resource.Quantity {
	return s.maxCpu
}

func (s *podMetricsSampler) MaxMem() *resource.Quantity {
	return s.maxMem
}

func (s *podMetricsSampler) MinCPU() *resource.Quantity {
	return s.minCpu
}

func (s *podMetricsSampler) MinMem() *resource.Quantity {
	return s.minMem
}

func (s *podMetricsSampler) Reset() {
	s.consumptionSamples = []PodMetricsEntry{}
}

func (s *podMetricsSampler) SampleConsumption(label string) error {
	podMetrics := &v1beta1.PodMetrics{}
	key := client.ObjectKey{Name: s.podName, Namespace: s.podNamespace}

	err := s.cl.Get(context.Background(), key, podMetrics)
	if err != nil {
		return err
	}
	if len(podMetrics.Containers) != 1 {
		return fmt.Errorf("expected one PodMetric for controller, but found %d", len(podMetrics.Containers))
	}
	s.consumptionSamples = append(s.consumptionSamples, PodMetricsEntry{Label: label, PodMetrics: podMetrics})

	if s.maxCpu == nil || podMetrics.Containers[0].Usage.Cpu().AsDec().Cmp(s.maxCpu.AsDec()) == 1 {
		s.maxCpu = podMetrics.Containers[0].Usage.Cpu()
	}
	if s.maxMem == nil || podMetrics.Containers[0].Usage.Memory().AsDec().Cmp(s.maxMem.AsDec()) == 1 {
		s.maxMem = podMetrics.Containers[0].Usage.Memory()
	}
	if s.minCpu == nil || podMetrics.Containers[0].Usage.Cpu().AsDec().Cmp(s.minCpu.AsDec()) == -1 {
		s.minCpu = podMetrics.Containers[0].Usage.Cpu()
	}
	if s.minMem == nil || podMetrics.Containers[0].Usage.Memory().AsDec().Cmp(s.minMem.AsDec()) == -1 {
		s.minMem = podMetrics.Containers[0].Usage.Memory()
	}

	fmt.Printf("\nMETRICS %s CPU:%v Mem:%v (%s)\n", time.Now().Format(time.RFC3339), podMetrics.Containers[0].Usage.Cpu(), podMetrics.Containers[0].Usage.Memory(), label)
	return nil
}
