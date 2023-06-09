package testing

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"reflect"
)

type CompareOptions struct {
	IgnoreMetadata       bool
	IgnoreOwnerRefs      bool
	IgnoreLabels         bool
	IgnoreMetadataFields []string
	CMPOption            CMPOption
}

type CMPOption func() (cmp.Options, error)

func (c *TemplateTestCase) stripIgnoredFields(expected *unstructured.Unstructured, actual *unstructured.Unstructured) {
	delete(expected.Object, "status")
	delete(actual.Object, "status")

	if c.CompareOptions != nil && c.CompareOptions.IgnoreLabels {
		expected.SetLabels(nil)
		actual.SetLabels(nil)
	}

	if c.CompareOptions != nil && c.CompareOptions.IgnoreMetadata {
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

	if c.CompareOptions != nil && c.CompareOptions.IgnoreOwnerRefs {
		delete(expectedMetadata, "ownerReferences")
		delete(actualMetadata, "ownerReferences")
	}

	if c.CompareOptions != nil {
		for _, field := range c.CompareOptions.IgnoreMetadataFields {
			delete(expectedMetadata, field)
			delete(actualMetadata, field)
		}
	}
}

func ConvertNumbersToFloatsDuringComparison() (cmp.Options, error) {
	return cmp.Options{
		cmp.FilterValues(func(x, y interface{}) bool {
			isNumeric := func(v interface{}) bool {
				return v != nil && reflect.TypeOf(v).ConvertibleTo(reflect.TypeOf(float64(0)))
			}
			return isNumeric(x) && isNumeric(y)
		}, cmp.Transformer("T", func(v interface{}) float64 {
			return reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Float()
		})),
	}, nil
}
