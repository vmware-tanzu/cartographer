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

package v1alpha1

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

func validateTemplate(tpl runtime.RawExtension) error {
	obj := metav1.PartialObjectMetadata{}
	if err := json.Unmarshal(tpl.Raw, &obj); err != nil {
		return fmt.Errorf("failed to parse object metadata: %w", err)
	}
	if obj.Namespace != metav1.NamespaceNone {
		return errors.New("template should not set metadata.namespace on the child object")
	}
	return nil
}
