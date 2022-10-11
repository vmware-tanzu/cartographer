package stamp

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type Reader interface {
	// fixme: output as a one-of is so weird
	GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error)
}

func NewReader(template client.Object, outputPaths map[string]string, inputGenerator realizer.Inputs) (Reader, error) {
	switch _ := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewSourceOutputReader(outputPaths), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewImageOutputReader(outputPaths), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewConfigOutputReader(outputPaths), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		return NewDeploymentPassThroughReader(template, inputGenerator), nil
	case *v1alpha1.ClusterTemplate:
		return NewNoOutputReader(), nil
	}
	return nil, fmt.Errorf("resource does not match a known template")
}

type SourceOutputReader struct{}

func (r *SourceOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func NewSourceOutputReader(paths map[string]string) Reader {
	return &SourceOutputReader{}
}

type ConfigOutputReader struct{}

func (r *ConfigOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func NewConfigOutputReader(paths map[string]string) Reader {
	return &ConfigOutputReader{}
}

type ImageOutputReader struct{}

func (r *ImageOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func NewImageOutputReader(paths map[string]string) Reader {
	return &ImageOutputReader{}
}

type DeploymentPassThroughReader struct{}

func (r *DeploymentPassThroughReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func NewDeploymentPassThroughReader(template client.Object, generator realizer.Inputs) Reader {
	return &DeploymentPassThroughReader{}
}

type NoOutputReader struct{}

func (r *NoOutputReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	return nil, fmt.Errorf("not implemented yet")

}

func NewNoOutputReader() Reader {
	return &NoOutputReader{}
}
