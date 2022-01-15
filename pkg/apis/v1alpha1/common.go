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

package v1alpha1

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TemplateParams []TemplateParam

type TemplateParam struct {
	Name         string               `json:"name"`
	DefaultValue apiextensionsv1.JSON `json:"default"`
}

type Param struct {
	Name  string               `json:"name"`
	Value apiextensionsv1.JSON `json:"value"`
}

type DelegatableParam struct {
	Name         string                `json:"name"`
	Value        *apiextensionsv1.JSON `json:"value,omitempty"`
	DefaultValue *apiextensionsv1.JSON `json:"default,omitempty"`
}

func (p *DelegatableParam) validateDelegatableParams() error {
	if p.bothValuesSet() || p.neitherValueSet() {
		return fmt.Errorf("param [%s] is invalid: must set exactly one of value and default", p.Name)
	}
	return nil
}

func (p *DelegatableParam) bothValuesSet() bool {
	return p.DefaultValue != nil && p.Value != nil
}

func (p *DelegatableParam) neitherValueSet() bool {
	return p.DefaultValue == nil && p.Value == nil
}

type ResourceReference struct {
	Name     string `json:"name"`
	Resource string `json:"resource"`
}

type Source struct {
	Git *GitSource `json:"git,omitempty"`
	// Image is an OCI image is a registry that contains source code
	Image   *string `json:"image,omitempty"`
	Subpath *string `json:"subPath,omitempty"`
}

type GitSource struct {
	URL *string `json:"url,omitempty"`
	Ref *GitRef `json:"ref,omitempty"`
}

type GitRef struct {
	Branch *string `json:"branch,omitempty"`
	Tag    *string `json:"tag,omitempty"`
	Commit *string `json:"commit,omitempty"`
}

type ObjectReference struct {
	Kind       string `json:"kind,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

type ServiceAccountRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

func GetAPITemplate(templateKind string) (client.Object, error) {
	var template client.Object

	switch templateKind {
	case "ClusterSourceTemplate":
		template = &ClusterSourceTemplate{}
	case "ClusterImageTemplate":
		template = &ClusterImageTemplate{}
	case "ClusterConfigTemplate":
		template = &ClusterConfigTemplate{}
	case "ClusterTemplate":
		template = &ClusterTemplate{}
	case "ClusterDeploymentTemplate":
		template = &ClusterDeploymentTemplate{}
	default:
		return nil, fmt.Errorf("resource does not have valid kind: %s", templateKind)
	}
	return template, nil
}

type Artifact struct {
	Id     string         `json:"id"`
	Passed PassedResource `json:"passed"`
	Source *SourceArtifact `json:"source,omitempty"`
	Image  *ImageArtifact  `json:"image,omitempty"`
	Config *ConfigArtifact `json:"config,omitempty"`
}

type SourceArtifact struct {
	Url      string `json:"url"`
	Revision string `json:"revision"`
}

type ImageArtifact struct {
	Image string `json:"image"`
}

type ConfigArtifact struct {
	Config string `json:"config"`
}

type PassedResource struct {
	ResourceName    string `json:"resource-name"`
	Kind            string `json:"kind"`
	ApiVersion      string `json:"apiVersion"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	ResourceVersion string `json:"resourceVersion"`
}
