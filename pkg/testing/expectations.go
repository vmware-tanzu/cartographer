package testing

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type Expectation interface {
	getExpected() (*unstructured.Unstructured, error)
}

type ExpectedFile struct {
	ExpectedFile string
}

type ExpectedUnstructured struct {
	ExpectedUnstructured *unstructured.Unstructured
}

func (e *ExpectedUnstructured) getExpected() (*unstructured.Unstructured, error) {
	return e.ExpectedUnstructured, nil
}

type ExpectedObject struct {
	ExpectedObject client.Object
}

func (e *ExpectedObject) getExpected() (*unstructured.Unstructured, error) {
	unstruct, err := runtime.DefaultUnstructuredConverter.ToUnstructured(e.ExpectedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert template to unstructured: %w", err)
	}

	return &unstructured.Unstructured{Object: unstruct}, nil
}

func (e *ExpectedFile) getExpected() (*unstructured.Unstructured, error) {
	expectedStampedObjectYaml, err := os.ReadFile(e.ExpectedFile)
	if err != nil {
		return nil, fmt.Errorf("could not read expected yaml: %w", err)
	}

	expectedJson, err := yaml.YAMLToJSON(expectedStampedObjectYaml)
	if err != nil {
		return nil, fmt.Errorf("convert yaml to json: %w", err)
	}

	expectedStampedObject := unstructured.Unstructured{}

	if err = expectedStampedObject.UnmarshalJSON(expectedJson); err != nil {
		return nil, fmt.Errorf("unmarshall json: %w", err)
	}

	return &expectedStampedObject, nil
}
