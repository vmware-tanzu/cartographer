package testing

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TemplateTestCase is an individual template test.
// Given and Expect values must be provided.
// Fields in the expected object's metadata may be ignored
// When run as part of a TemplateTestSuite, an individual case(s) may be focused.
// This will exercise the individual test(s).
// Note that the overall suite will fail (preventing focused tests from passing CI).
type TemplateTestCase struct {
	Given                TemplateTestGivens
	Expect               TemplateTestExpectation
	IgnoreMetadata       bool
	IgnoreOwnerRefs      bool
	IgnoreLabels         bool
	IgnoreMetadataFields []string
	Focus                bool
}

func (c *TemplateTestCase) Run() error {
	expectedObject, err := c.Expect.getExpectedObject()
	if err != nil {
		return fmt.Errorf("failed to get expected object: %w", err)
	}

	actualObject, err := c.Given.getActualObject()
	if errors.Is(err, yttNotFound) {
		return fmt.Errorf("test requires ytt, but ytt was not found in path")
	} else if err != nil {
		return fmt.Errorf("failed to get actual object: %w", err)
	}

	c.stripIgnoredFields(expectedObject, actualObject)

	if diff := cmp.Diff(expectedObject.Object, actualObject.Object); diff != "" {
		return fmt.Errorf("expected does not equal actual: (-expected +actual):\n%s", diff)
	}

	return nil
}

func (c *TemplateTestCase) stripIgnoredFields(expected *unstructured.Unstructured, actual *unstructured.Unstructured) {
	delete(expected.Object, "status")
	delete(actual.Object, "status")

	if c.IgnoreLabels {
		expected.SetLabels(nil)
		actual.SetLabels(nil)
	}

	if c.IgnoreMetadata {
		delete(expected.Object, "metadata")
		delete(actual.Object, "metadata")
	}

	var expectedMetadata, actualMetadata map[string]interface{}

	if expected.Object["metadata"] != nil {
		expectedMetadata = expected.Object["metadata"].(map[string]interface{})
	}
	if actual.Object["metadata"] != nil {
		actualMetadata = actual.Object["metadata"].(map[string]interface{})
	}

	if c.IgnoreOwnerRefs {
		delete(expectedMetadata, "ownerReferences")
		delete(actualMetadata, "ownerReferences")
	}

	for _, field := range c.IgnoreMetadataFields {
		delete(expectedMetadata, field)
		delete(actualMetadata, field)
	}
}

func (i *TemplateTestGivens) getActualObject() (*unstructured.Unstructured, error) {
	ctx := context.Background()

	workload, err := i.Workload.GetWorkload()
	if err != nil {
		return nil, fmt.Errorf("get workload failed: %w", err)
	}

	apiTemplate, err := i.Template.GetTemplate()
	if err != nil {
		return nil, fmt.Errorf("get populated template failed: %w", err)
	}

	if err = (*apiTemplate).ValidateCreate(); err != nil {
		return nil, fmt.Errorf("template validation failed: %w", err)
	}

	template, err := templates.NewReaderFromAPI(*apiTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster template")
	}

	if template.IsYTTTemplate() {
		err = ensureYTTAvailable(ctx)
		if err != nil {
			return nil, fmt.Errorf("ensure YTT available: %w", err)
		}
	}

	if i.actualBlueprintSupplied() {
		return i.actualBlueprintStamp(ctx, workload, template)
	}

	return i.MockSupplyChain.mockedBlueprintStamp(ctx, workload, *apiTemplate, template)
}
