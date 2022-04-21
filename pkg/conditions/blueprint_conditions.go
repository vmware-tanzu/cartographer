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

package conditions

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

// -- Blueprint.Status.Conditions - TemplatesReady

func TemplatesNotFoundCondition(resourceNames []string) metav1.Condition {
	message := fmt.Sprintf(
		"did not find the template of the resource(s) [%s]",
		strings.Join(resourceNames, ", "),
	)

	return metav1.Condition{
		Type:    v1alpha1.BlueprintTemplatesReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.NotFoundTemplatesReadyReason,
		Message: message,
	}
}

func TemplatesFoundCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.BlueprintTemplatesReady,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.ReadyTemplatesReadyReason,
	}
}
