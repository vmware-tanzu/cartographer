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

package templates

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type Reader interface {
	GetDefaultParams() v1alpha1.TemplateParams

	// GetResourceTemplate returns the actual representation of a resource to stamp, and how to handle it
	// TODO: we should be expecting something with a [ytt|template] interface, the health rules and params should
	// not be fetched here
	GetResourceTemplate() v1alpha1.TemplateSpec
	GetHealthRule() *v1alpha1.HealthRule
	IsYTTTemplate() bool
	GetLifecycle() *Lifecycle
	GetRetentionPolicy() v1alpha1.RetentionPolicy
}

type Lifecycle string

const (
	Default   Lifecycle = "default"
	Immutable Lifecycle = "immutable"
	Tekton    Lifecycle = "tekton"
)

func (l *Lifecycle) IsImmutable() bool {
	if *l == Immutable || *l == Tekton {
		return true
	}
	return false
}

func convertLifecycle(lifecycleString string) Lifecycle {
	switch lifecycleString {
	case "immutable":
		return Immutable
	case "tekton":
		return Tekton
	default:
		return Default
	}
}

func NewReaderFromAPI(template client.Object) (Reader, error) {
	switch v := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewClusterSourceTemplateReader(v), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewClusterImageTemplateReader(v), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewClusterConfigTemplateReader(v), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		return NewClusterDeploymentTemplateReader(v), nil
	case *v1alpha1.ClusterTemplate:
		return NewClusterTemplateReader(v), nil
	}
	return nil, fmt.Errorf("resource does not match a known template")
}
