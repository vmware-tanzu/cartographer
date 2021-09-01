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

package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/valyala/fasttemplate"
)

type TemplateExecutor func(template, startTag, endTag string, f fasttemplate.TagFunc) (string, error)

//counterfeiter:generate . evaluator
type evaluator interface {
	EvaluateJsonPath(path string, obj interface{}) (interface{}, error)
}

//counterfeiter:generate . tagInterpolator
type tagInterpolator interface {
	InterpolateTag(w io.Writer, tag string) (int, error)
	Evaluate(tag string) (interface{}, error)
}

func isSingleTag(template string) bool {
	return strings.HasPrefix(template, `$(`) &&
		strings.HasSuffix(template, `)$`) &&
		strings.Count(template, `$(`) == 1
}

// InterpolateLeafNode merges the context variables anywhere a $(<<jsonPath>>)$ tag is found
// It validates that the jsonPath refers to objects within the context
// It also makes workloads easier to access by aliasing the content of 'metadata' and 'spec' keys
//   onto 'workload.'
func InterpolateLeafNode(executor TemplateExecutor, template []byte, tagInterpolator tagInterpolator) (interface{}, error) {
	input := string(template)

	if isSingleTag(input) {
		jsonPathExpr := strings.TrimPrefix(strings.TrimSuffix(input, `)$`), `$(`)
		result, err := tagInterpolator.Evaluate(jsonPathExpr)

		if err != nil {
			return nil, fmt.Errorf("evaluate tag %s: %w", input, err)
		}
		return result, nil
	}

	stringResult, err := executor(input, `$(`, `)$`, tagInterpolator.InterpolateTag)
	if err != nil {
		return nil, fmt.Errorf("interpolate tag: %w", err)
	}

	return stringResult, nil
}

type StandardTagInterpolator struct {
	Context   StampContext
	Evaluator evaluator
}

//counterfeiter:generate io.Writer
func (t StandardTagInterpolator) Evaluate(tag string) (interface{}, error) {
	return t.Evaluator.EvaluateJsonPath(tag, t.Context)
}

func (t StandardTagInterpolator) InterpolateTag(w io.Writer, tag string) (int, error) {
	var (
		val       interface{}
		err       error
		writeLen  int
		jsonValue []byte
	)

	val, err = t.Evaluator.EvaluateJsonPath(tag, t.Context)
	if err != nil {
		return 0, fmt.Errorf("evaluate jsonpath: %w", err)
	}

	if val == nil {
		return 0, fmt.Errorf("tag must not point to nil value: %s", tag)
	}

	if reflect.TypeOf(val).Kind() == reflect.String {
		writeLen, err = w.Write([]byte(fmt.Sprintf("%s", val)))
	} else {
		jsonValue, err = json.Marshal(val)
		if err != nil {
			return 0, fmt.Errorf("json marshal: %w", err)
		}
		writeLen, err = w.Write(jsonValue)
	}

	if err != nil {
		return 0, fmt.Errorf("writer write: %w", err)
	}
	return writeLen, nil
}
