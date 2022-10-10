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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
)

const PREVIEW_CHARACTER_LIMIT = 1024

type Inputs interface {
	GetDeployment() *SourceInput
}

// Note: to become the interface that component consumes to describe the template loaded from etcd
// READ Oriented
type Reader interface {
	GetDefaultParams() v1alpha1.TemplateParams

	// GetResourceTemplate returns the actual representation of a resource to stamp, and how to handle it
	// TODO: we should be expecting something with a [ytt|template] interface, the health rules and params should
	// not be fetched here
	GetResourceTemplate() v1alpha1.TemplateSpec
	GetHealthRule() *v1alpha1.HealthRule

	GetOutputPath()

	// TODO Remove to it's own interface
	GenerateResourceOutput(output *Output) ([]v1alpha1.Output, error)
}

// Ephemeral results of stamping shit out
type StampedResult interface {
	//SetInputs(inputs Inputs)
	//GetOutput(stampedObject *unstructured.Unstructured) (*Output, error)
	GetOutput(stampedObject *unstructured.Unstructured, inputs Inputs) (*Output, error)
}

func NewReaderFromAPI(template client.Object) (Reader, error) {
	switch v := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewClusterSourceTemplateModel(v, eval.EvaluatorBuilder()), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewClusterImageTemplateModel(v, eval.EvaluatorBuilder()), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewClusterConfigTemplateModel(v, eval.EvaluatorBuilder()), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		return NewClusterDeploymentTemplateModel(v, eval.EvaluatorBuilder()), nil
	case *v1alpha1.ClusterTemplate:
		return NewClusterTemplateModel(v), nil
	}
	return nil, fmt.Errorf("resource does not match a known template")
}
