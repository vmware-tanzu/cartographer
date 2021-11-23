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
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientBuilder func(secret *corev1.Secret) (client.Client, error)

func NewClientBuilder(restConfig *rest.Config) ClientBuilder {
	return func(secret *corev1.Secret) (client.Client, error) {
		config, err := AddBearerToken(secret, restConfig)
		if err != nil {
			return nil, fmt.Errorf("adding bearer token: %w", err)
		}

		cl, err := client.New(config, client.Options{})
		if err != nil {
			return nil, fmt.Errorf("creating client: %w", err)
		}

		return cl, nil
	}
}

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
