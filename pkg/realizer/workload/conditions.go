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

package workload

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

// -- NEW

func ResourceSubmittedCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.ResourceSubmitted,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.CompleteResourcesSubmittedReason,
	}
}

// -- OLD, LIFTED AND SHIFTED FROM controller/workload/conditions
// -- DO WE WANT TO USE THE SAME CONDITIONS WE CURRENTLY HAVE? WILL NEED TO MOVE TO NEW PACKAGE - conditions/workload

func TemplateObjectRetrievalFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateObjectRetrievalFailureResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func MissingValueAtPathCondition(obj *unstructured.Unstructured, expression string) metav1.Condition {
	var namespaceMsg string
	if obj.GetNamespace() != "" {
		namespaceMsg = fmt.Sprintf(" in namespace [%s]", obj.GetNamespace())
	}
	return metav1.Condition{
		Type:   v1alpha1.WorkloadResourceSubmitted,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.MissingValueAtPathResourcesSubmittedReason,
		Message: fmt.Sprintf("waiting to read value [%s] from resource [%s/%s]%s",
			expression, utils.GetFullyQualifiedType(obj), obj.GetName(), namespaceMsg),
	}
}

func TemplateStampFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func TemplateRejectedByAPIServerCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateRejectedByAPIServerResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func UnknownResourceErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.UnknownErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func ServiceAccountSecretNotFoundCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ServiceAccountSecretErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func ResourceRealizerBuilderErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ResourceRealizerBuilderErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func ResolveTemplateOptionsErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ResolveTemplateOptionsErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func TemplateOptionsMatchErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadResourceSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateOptionsMatchErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}
