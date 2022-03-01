package supplychains

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterSupplyChain struct {
	supplyChain *v1alpha1.ClusterSupplyChain
}

func (t *clusterSupplyChain) GetName() string {
	return t.supplyChain.GetName()
}

func (t *clusterSupplyChain) GetResources() []v1alpha1.SupplyChainResource {
	return t.supplyChain.Spec.Resources
}

func (t *clusterSupplyChain) GetOutputResource() string {
	return ""
}

func (t *clusterSupplyChain) GetParams() []v1alpha1.BlueprintParam {
	return t.supplyChain.Spec.Params
}

func (t *clusterSupplyChain) GetStatus() v1alpha1.SupplyChainStatus {
	return t.supplyChain.Status
}

func (t *clusterSupplyChain) GetServiceAccountRef() v1alpha1.ServiceAccountRef {
	return t.supplyChain.Spec.ServiceAccountRef
}

func NewClusterSupplyChain(supplyChain *v1alpha1.ClusterSupplyChain) *clusterSupplyChain {
	return &clusterSupplyChain{supplyChain: supplyChain}
}
