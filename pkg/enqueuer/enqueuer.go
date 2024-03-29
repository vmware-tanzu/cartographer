/*
Copyright 2020 VMware, Inc.
SPDX-License-Identifier: Apache-2.0
*/
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

package enqueuer

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	tracker "github.com/vmware-tanzu/cartographer/pkg/tracker/dependency"
)

func EnqueueTracked(by client.Object, t tracker.DependencyTracker, s *runtime.Scheme) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(
		func(_ context.Context, a client.Object) []reconcile.Request {
			var requests []reconcile.Request

			gvks, _, err := s.ObjectKinds(by)
			if err != nil {
				panic(err)
			}

			key := tracker.NewKey(
				gvks[0],
				types.NamespacedName{Namespace: a.GetNamespace(), Name: a.GetName()},
			)
			for _, item := range t.Lookup(key) {
				requests = append(requests, reconcile.Request{NamespacedName: item})
			}

			return requests
		},
	)
}
