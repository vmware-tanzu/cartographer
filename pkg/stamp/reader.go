package stamp

import (
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/eval"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

type DeploymentInput interface {
	GetDeployment() *templates.SourceInput
}

type Reader interface {
	// fixme: output as a one-of is so weird
	GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error)
}

func NewReader(template client.Object, inputReader DeploymentInput) (Reader, error) {
	switch v := template.(type) {

	case *v1alpha1.ClusterSourceTemplate:
		return NewSourceOutputReader(v), nil
	case *v1alpha1.ClusterImageTemplate:
		return NewImageOutputReader(v), nil
	case *v1alpha1.ClusterConfigTemplate:
		return NewConfigOutputReader(v), nil
	case *v1alpha1.ClusterDeploymentTemplate:
		// TODO: we don't need the template, right??
		return NewDeploymentPassThroughReader(inputReader), nil
	case *v1alpha1.ClusterTemplate:
		return NewNoOutputReader(), nil
	}
	return nil, fmt.Errorf("resource does not match a known template")
}

type SourceOutputReader struct {
	template *v1alpha1.ClusterSourceTemplate
}

func (r *SourceOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	// TODO: We don't need a Builder
	evaluator := eval.EvaluatorBuilder()
	url, err := evaluator.EvaluateJsonPath(r.template.Spec.URLPath, stampedObject.UnstructuredContent())
	if err != nil {
		// TODO: errors
		//return nil, JsonPathError{
		//	Err: fmt.Errorf("failed to evaluate the url path [%s]: %w",
		//		r.template.Spec.URLPath, err),
		//	expression: r.template.Spec.URLPath,
		//}
		return nil, err
	}

	revision, err := evaluator.EvaluateJsonPath(r.template.Spec.RevisionPath, stampedObject.UnstructuredContent())
	if err != nil {
		//return nil, JsonPathError{
		//	Err: fmt.Errorf("failed to evaluate the revision path [%s]: %w",
		//		r.template.Spec.RevisionPath, err),
		//	expression: r.template.Spec.RevisionPath,
		//}
		return nil, err
	}
	return &templates.Output{
		Source: &templates.Source{
			URL:      url,
			Revision: revision,
		},
	}, nil
	return nil, fmt.Errorf("not implemented yet")
}

func NewSourceOutputReader(template *v1alpha1.ClusterSourceTemplate) Reader {
	return &SourceOutputReader{template: template}
}

type ConfigOutputReader struct {
	template *v1alpha1.ClusterConfigTemplate
}

func (r *ConfigOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	evaluator := eval.EvaluatorBuilder()
	config, err := evaluator.EvaluateJsonPath(r.template.Spec.ConfigPath, stampedObject.UnstructuredContent())
	if err != nil {
		//return nil, JsonPathError{
		//	Err: fmt.Errorf("failed to evaluate spec.configPath [%s]: %w",
		//		r.template.Spec.ConfigPath, err),
		//	expression: r.template.Spec.ConfigPath,
		//}
		return nil, err
	}

	return &templates.Output{
		Config: config,
	}, nil
}

func NewConfigOutputReader(template *v1alpha1.ClusterConfigTemplate) Reader {
	return &ConfigOutputReader{
		template: template,
	}
}

type ImageOutputReader struct {
	template *v1alpha1.ClusterImageTemplate
}

func (r *ImageOutputReader) GetOutput(stampedObject *unstructured.Unstructured) (*templates.Output, error) {
	evaluator := eval.EvaluatorBuilder()
	image, err := evaluator.EvaluateJsonPath(r.template.Spec.ImagePath, stampedObject.UnstructuredContent())
	if err != nil {
		//return nil, JsonPathError{
		//	Err: fmt.Errorf("failed to evaluate the url path [%s]: %w",
		//		r.template.Spec.ImagePath, err),
		//	expression: r.template.Spec.ImagePath,
		//}
		return nil, err
	}

	return &templates.Output{
		Image: image,
	}, nil
}

func NewImageOutputReader(template *v1alpha1.ClusterImageTemplate) Reader {
	return &ImageOutputReader{
		template: template,
	}
}

type DeploymentPassThroughReader struct {
	inputs DeploymentInput
}

func (r *DeploymentPassThroughReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	// TODO: go get the other private funcs
	//if err := t.outputReady(stampedObject); err != nil {
	//	return nil, err
	//}

	output := &templates.Output{Source: &templates.Source{}}

	deployment := r.inputs.GetDeployment()
	if deployment == nil {
		return nil, fmt.Errorf("deployment not found in upstream template")
	}

	output.Source.URL = deployment.URL
	output.Source.Revision = deployment.Revision

	return output, nil
}

func NewDeploymentPassThroughReader(inputReader DeploymentInput) Reader {
	return &DeploymentPassThroughReader{
		inputs: inputReader,
	}
}

type NoOutputReader struct{}

func (r *NoOutputReader) GetOutput(_ *unstructured.Unstructured) (*templates.Output, error) {
	return &templates.Output{}, nil

}

func NewNoOutputReader() Reader {
	return &NoOutputReader{}
}
