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

package registrar

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

//counterfeiter:generate . Logger
type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
}

type Mapper struct {
	Client client.Client
	Logger Logger
}

func (mapper *Mapper) ClusterSupplyChainToWorkloadRequests(object client.Object) []reconcile.Request {
	var err error

	supplyChain, ok := object.(*v1alpha1.ClusterSupplyChain)
	if !ok {
		mapper.Logger.Error(nil, "cluster supply chain to workload requests: cast to ClusterSupplyChain failed")
		return nil
	}

	list := &v1alpha1.WorkloadList{}

	err = mapper.Client.List(context.TODO(), list,
		client.InNamespace(supplyChain.Namespace),
		client.MatchingLabels(supplyChain.Spec.Selector))
	if err != nil {
		mapper.Logger.Error(fmt.Errorf("client list: %w", err), "cluster supply chain to workload requests: client list")
		return nil
	}

	var requests []reconcile.Request
	for _, workload := range list.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      workload.Name,
				Namespace: workload.Namespace,
			},
		})
	}

	return requests

}
