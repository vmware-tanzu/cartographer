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
