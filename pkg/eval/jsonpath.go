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

package eval

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

//counterfeiter:generate . Evaluate
type Evaluate func(jsonpathExpression string, obj interface{}) ([]interface{}, error)

type Evaluator struct {
	Evaluate Evaluate
}

func EvaluatorBuilder() Evaluator {
	return Evaluator{
		Evaluate: utils.SinglePathEvaluate,
	}
}

func (e Evaluator) EvaluateJsonPath(path string, obj interface{}) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("empty jsonpath not allowed")
	}

	jsonpathExpression := ensureValidWrapping(path)

	interfaceList, err := e.Evaluate(jsonpathExpression, obj)
	if err != nil {
		return nil, fmt.Errorf("evaluate: %w", err)
	}

	if len(interfaceList) > 1 {
		return "", fmt.Errorf("too many results for the query: %s", path)
	}

	if len(interfaceList) == 0 {
		return "", fmt.Errorf("jsonpath returned empty list: %s", path)
	}

	return interfaceList[0], nil
}

func ensureValidWrapping(jsonpathExpression string) string {
	if !strings.HasPrefix(jsonpathExpression, "{.") {
		if !strings.HasPrefix(jsonpathExpression, ".") {
			jsonpathExpression = fmt.Sprintf("{.%s", jsonpathExpression)
		} else {
			jsonpathExpression = fmt.Sprintf("{%s", jsonpathExpression)
		}
	}

	if !strings.HasSuffix(jsonpathExpression, "}") {
		jsonpathExpression = fmt.Sprintf("%s}", jsonpathExpression)
	}

	return jsonpathExpression
}
