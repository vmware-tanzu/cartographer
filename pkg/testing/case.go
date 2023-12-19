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
	"context"
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

// Test is an individual template test.
// Given and Expect values must be provided.
// Fields in the expected object's metadata may be ignored
// When run as part of a Suite, an individual case(s) may be focused.
// This will exercise the individual test(s).
// Note that the overall suite will fail (preventing focused tests from passing CI).
type Test struct {
	Given          Given
	Expect         Expectation
	CompareOptions *CompareOptions
	Focus          bool
}

// Given must specify a Template and a Workload.
// SupplyChain is optional
type Given struct {
	Template    Template
	Workload    Workload
	SupplyChain SupplyChain
}

func (c *Test) Run() error {
	expectedObject, err := c.Expect.getExpected()
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

	var opts cmp.Options
	if c.CompareOptions != nil && c.CompareOptions.CMPOption != nil {
		opts, err = c.CompareOptions.CMPOption()
		if err != nil {
			return fmt.Errorf("get compare options: %w", err)
		}
	}

	if diff := cmp.Diff(expectedObject.Object, actualObject.Object, opts); diff != "" {
		return fmt.Errorf("expected does not equal actual: (-expected +actual):\n%s", diff)
	}

	return nil
}

func (i *Given) getActualObject() (*unstructured.Unstructured, error) {
	ctx := context.Background()

	workload, err := i.Workload.GetWorkload()
	if err != nil {
		return nil, fmt.Errorf("get workload failed: %w", err)
	}

	apiTemplate, err := i.Template.GetTemplate()
	if err != nil {
		return nil, fmt.Errorf("get populated template failed: %w", err)
	}

	if _, err = (*apiTemplate).ValidateCreate(); err != nil {
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

	if i.SupplyChain == nil {
		i.SupplyChain = &MockSupplyChain{}
	}

	return i.SupplyChain.stamp(ctx, workload, *apiTemplate, template)
}
