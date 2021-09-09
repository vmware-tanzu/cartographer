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

package workload

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(componentRealizer ComponentRealizer, supplyChain *v1alpha1.ClusterSupplyChain) error
}

type realizer struct{}

func NewRealizer() Realizer {
	return &realizer{}
}

func (r *realizer) Realize(componentRealizer ComponentRealizer, supplyChain *v1alpha1.ClusterSupplyChain) error {
	outs := NewOutputs()

	for i := range supplyChain.Spec.Components {
		component := supplyChain.Spec.Components[i]
		out, err := componentRealizer.Do(&component, supplyChain.Name, outs)
		if err != nil {
			return err
		}
		outs.AddOutput(component.Name, out)
	}

	return nil
}
