package supplychains

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterSupplyChain struct {
	supplyChain *v1alpha1.ClusterSupplyChain
}

func (t *clusterSupplyChain) GetName() string {
	return t.GetName()
}

func (t *clusterSupplyChain) GetResources() []v1alpha1.SupplyChainResource {
	return t.supplyChain.Spec.Resources
}

func (t *clusterSupplyChain) GetOutputResource() string {
	return ""
}

func NewClusterSupplyChain(supplyChain *v1alpha1.ClusterSupplyChain) *clusterSupplyChain {
	return &clusterSupplyChain{supplyChain: supplyChain}
}
