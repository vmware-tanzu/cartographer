// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testing

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CompareOptions struct {
	IgnoreMetadata       bool
	IgnoreOwnerRefs      bool
	IgnoreLabels         bool
	IgnoreMetadataFields []string
	CMPOption            CMPOption
}

type CMPOption func() (cmp.Options, error)

func (c *Test) stripIgnoredFields(expected *unstructured.Unstructured, actual *unstructured.Unstructured) {
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
