package supplychains

import (
	"strings"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type clusterSourceSupplyChain struct {
	supplyChain *v1alpha1.ClusterSourceSupplyChain
}

func (t *clusterSourceSupplyChain) GetName() string {
	return t.supplyChain.GetName()
}

func (t *clusterSourceSupplyChain) GetResources() []v1alpha1.SupplyChainResource {
	return t.supplyChain.Spec.Resources
}

func (t *clusterSourceSupplyChain) GetOutputResource() string {
	//TODO this makes assumption that url and revision need to come from same resource
	return strings.Split(t.supplyChain.Spec.URLPath, ".")[0]
}

func (t *clusterSourceSupplyChain) GetParams() []v1alpha1.BlueprintParam {
	return t.supplyChain.Spec.Params
}

func (t *clusterSourceSupplyChain) GetStatus() v1alpha1.SupplyChainStatus {
	return t.supplyChain.Status
}

func (t *clusterSourceSupplyChain) GetServiceAccountRef() v1alpha1.ServiceAccountRef {
	return t.supplyChain.Spec.ServiceAccountRef
}

func NewClusterSourceSupplyChain(supplyChain *v1alpha1.ClusterSourceSupplyChain) *clusterSourceSupplyChain {
	return &clusterSourceSupplyChain{supplyChain: supplyChain}
}
