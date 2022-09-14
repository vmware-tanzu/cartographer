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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const DefaultResyncTime = 10 * time.Hour

func GetObjectGVK(obj metav1.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	ro, ok := obj.(runtime.Object)
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("%T is not a runtime.Object", obj)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("gvk for object: %w", err)
	}

	return gvk, nil
}

func HereYaml(y string) string {
	y = strings.ReplaceAll(y, "\t", "    ")
	return heredoc.Doc(y)
}

func HereYamlF(y string, args ...interface{}) string {
	y = strings.ReplaceAll(y, "\t", "    ")
	return heredoc.Docf(y, args...)
}

func AlterFieldOfNestedStringMaps(obj interface{}, key string, value string) error {
	switch knownType := obj.(type) {
	case map[string]interface{}:
		i := strings.Index(key, ".")
		if i < 0 {
			_, ok := knownType[key]
			if !ok {
				return errors.New("field not found")
			}
			knownType[key] = value
		} else {
			keyPrefix := key[:i]
			keySuffix := key[i+1:]
			subMap, ok := knownType[keyPrefix]
			if !ok {
				return errors.New("field not found")
			}
			return AlterFieldOfNestedStringMaps(subMap, keySuffix, value)
		}

		return nil
	case []interface{}:
		if key[:1] != "[" {
			return errors.New("field not found")
		}
		closeBracket := strings.Index(key, "]")
		if closeBracket < 0 {
			return errors.New("field not found")
		}
		strIndex := key[1:closeBracket]
		intIndex, err := strconv.Atoi(strIndex)
		if err != nil {
			return errors.New("field not found")
		}
		keySuffix := key[closeBracket+1:]
		return AlterFieldOfNestedStringMaps(knownType[intIndex], keySuffix, value)
	default:
		return fmt.Errorf("unexpected type: %T", knownType)
	}
}

func GetFullyQualifiedType(obj *unstructured.Unstructured) string {
	var fullyQualifiedType string
	if obj.GetObjectKind().GroupVersionKind().Group == "" {
		fullyQualifiedType = strings.ToLower(obj.GetKind())
	} else {
		fullyQualifiedType = fmt.Sprintf("%s.%s", strings.ToLower(obj.GetKind()),
			obj.GetObjectKind().GroupVersionKind().Group)
	}
	return fullyQualifiedType
}
