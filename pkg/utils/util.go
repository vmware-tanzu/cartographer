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
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

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

