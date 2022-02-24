package supplychains

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type SupplyChain interface {
	GetResources() []v1alpha1.SupplyChainResource
	GetName() string
	GetOutputResource() string
	GetParams() []v1alpha1.BlueprintParam
	GetStatus() v1alpha1.SupplyChainStatus
	GetServiceAccountRef() v1alpha1.ServiceAccountRef
}

func NewModelFromAPI(supplyChain client.Object) (SupplyChain, error) {
	switch v := supplyChain.(type) {

	case *v1alpha1.ClusterSupplyChain:
		return NewClusterSupplyChain(v), nil
	case *v1alpha1.ClusterSourceSupplyChain:
		return NewClusterSourceSupplyChain(v), nil
	}
	return nil, fmt.Errorf("resource does not match a known supply chain")
}
