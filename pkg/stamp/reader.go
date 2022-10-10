package stamp

import (
	"fmt"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reader interface {
	GetOutput(stampedObject *unstructured.Unstructured) (*Output, error)
	GetPassthroughOutput() (*Output, error)
}

func NewReader(template client.Object, outputPaths map[string]string, inputGenerator realizer.Inputs) (Reader, error) {
	switch v := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewClusterSourceOutputReader(outputPaths, passthrough), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewClusterImageOutputReader(outputPaths, passthrough), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewClusterConfigOutputReader(outputPaths, passthrough), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		return NewClusterDeploymentOutputReader(inputGenerator, passthrough), nil
	case *v1alpha1.ClusterTemplate:
		return NewNoOutputReader(), nil
	}
	return nil, fmt.Errorf("resource does not match a known template")
}
