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

package realizer

import "github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"

type OwnerResource struct {
	TemplateRef     v1alpha1.TemplateReference
	TemplateOptions []v1alpha1.TemplateOption
	Params          []v1alpha1.BlueprintParam
	Name            string
	Sources         []v1alpha1.ResourceReference
	Images          []v1alpha1.ResourceReference
	Configs         []v1alpha1.ResourceReference
	Deployment      *v1alpha1.DeploymentReference
}

func (o OwnerResource) GetImages() []v1alpha1.ResourceReference {
	return o.Images
}

func (o OwnerResource) GetConfigs() []v1alpha1.ResourceReference {
	return o.Configs
}

func (o OwnerResource) GetDeployment() *v1alpha1.DeploymentReference {
	return o.Deployment
}

func (o OwnerResource) GetName() string {
	return o.Name
}

func (o OwnerResource) GetSources() []v1alpha1.ResourceReference {
	return o.Sources
}
