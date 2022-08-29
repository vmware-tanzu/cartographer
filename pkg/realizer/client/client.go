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

package client

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientBuilder func(authToken string, needDiscovery bool) (client.Client, discovery.DiscoveryInterface, error)

func NewClientBuilder(restConfig *rest.Config) ClientBuilder {
	return func(authToken string, needDiscovery bool) (client.Client, discovery.DiscoveryInterface, error) {
		newConfig := *restConfig
		newConfig.BearerToken = authToken
		newConfig.BearerTokenFile = ""

		cl, err := client.New(&newConfig, client.Options{})
		if err != nil {
			return nil, nil, fmt.Errorf("creating client: %w", err)
		}

		var cachedDiscoveryClient discovery.DiscoveryInterface
		if needDiscovery {
			discoveryClient, err := discovery.NewDiscoveryClientForConfig(&newConfig)
			if err != nil {
				return cl, nil, fmt.Errorf("failed to create discovery client: %w", err)
			}
			cachedDiscoveryClient = memory.NewMemCacheClient(discoveryClient)
		}

		return cl, cachedDiscoveryClient, nil
	}
}

//TODO: this must be removed --- compare what is tested in client_test ?
func AddBearerToken(secret *corev1.Secret, restConfig *rest.Config) (*rest.Config, error) {
	tokenBytes, found := secret.Data[corev1.ServiceAccountTokenKey]
	if !found {
		return nil, fmt.Errorf("couldn't find service account token value")
	}

	newConfig := *restConfig
	newConfig.BearerToken = string(tokenBytes)
	newConfig.BearerTokenFile = ""

	return &newConfig, nil
}
