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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/util/jsonpath"
)

func SinglePathEvaluate(jsonpathExpression string, obj interface{}) ([]interface{}, error) {
	var (
		jsonBuffer    bytes.Buffer
		interfaceList []interface{}
	)

	parser := jsonpath.New("")
	parser.AllowMissingKeys(true)

	err := parser.Parse(jsonpathExpression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jsonpath '%s': %w", jsonpathExpression, err)
	}

	values, err := parser.FindResults(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to find results: %w", err)
	}

	if len(values) > 1 {
		return nil, fmt.Errorf("more queries than expected: %v", values)
	}

	parser.EnableJSONOutput(true)
	err = parser.PrintResults(&jsonBuffer, values[0])
	if err != nil {
		return nil, fmt.Errorf("failed to print results: %w", err)
	}

	bufBytes := jsonBuffer.Bytes()
	err = json.Unmarshal(bufBytes, &interfaceList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall: %w", err)
	}
	return interfaceList, nil
}
