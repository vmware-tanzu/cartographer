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

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

// -- Supply Chain conditions

func SupplyChainReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.WorkloadSupplyChainReady,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.ReadySupplyChainReason,
	}
}

func WorkloadMissingLabelsCondition() metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadSupplyChainReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.WorkloadLabelsMissingSupplyChainReason,
		Message: "workload has no labels to match to supply chain",
	}
}

func SupplyChainNotFoundCondition(labels map[string]string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadSupplyChainReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.NotFoundSupplyChainReadyReason,
		Message: fmt.Sprintf("no supply chain found where full selector is satisfied by labels: %+v", labels),
	}
}

func TooManySupplyChainMatchesCondition() metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadSupplyChainReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.MultipleMatchesSupplyChainReadyReason,
		Message: "workload may only match a single supply chain's selector",
	}
}

func MissingReadyInSupplyChainCondition(supplyChainReadyCondition metav1.Condition) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadSupplyChainReady,
		Status:  metav1.ConditionFalse,
		Reason:  supplyChainReadyCondition.Reason,
		Message: supplyChainReadyCondition.Message,
	}
}

// -- Component conditions

func ComponentsSubmittedCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.WorkloadComponentsSubmitted,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.CompleteComponentsSubmittedReason,
	}
}

func TemplateObjectRetrievalFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadComponentsSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateObjectRetrievalFailureComponentsSubmittedReason,
		Message: err.Error(),
	}
}

func MissingValueAtPathCondition(componentName, expression string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadComponentsSubmitted,
		Status:  metav1.ConditionUnknown,
		Reason:  v1alpha1.MissingValueAtPathComponentsSubmittedReason,
		Message: fmt.Sprintf("Component '%s' is waiting to read value '%s'", componentName, expression),
	}
}

func TemplateStampFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadComponentsSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureComponentsSubmittedReason,
		Message: err.Error(),
	}
}

func TemplateRejectedByAPIServerCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadComponentsSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateRejectedByAPIServerComponentsSubmittedReason,
		Message: err.Error(),
	}
}

func UnknownComponentErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadComponentsSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.UnknownErrorComponentsSubmittedReason,
		Message: err.Error(),
	}
}
