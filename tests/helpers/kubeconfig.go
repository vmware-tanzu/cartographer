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

package helpers

import (
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func GenerateConfigFile(env *envtest.Environment) (string, error) {
	user, err := env.ControlPlane.AddUser(envtest.User{
		Name:   "envtest-admin",
		Groups: []string{"system:masters"}, // wokeignore:rule=master K8s is riddled with this term
	}, nil)
	if err != nil {
		return "", fmt.Errorf("add user: %w", err)
	}

	kubeconfigFile, err := os.CreateTemp("", "cartographer-integration-test-kubeconfig-")
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
