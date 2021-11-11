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

package deliverable

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/deliverable"
)

// -- Delivery conditions

func DeliveryReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.DeliverableDeliveryReady,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.ReadyDeliveryReason,
	}
}

func DeliverableMissingLabelsCondition() metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableDeliveryReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.DeliverableLabelsMissingDeliveryReason,
		Message: "deliverable has no labels to match to delivery",
	}
}

func DeliveryNotFoundCondition(labels map[string]string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableDeliveryReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.NotFoundDeliveryReadyReason,
		Message: fmt.Sprintf("no delivery found where full selector is satisfied by labels: %+v", labels),
	}
}

func TooManyDeliveryMatchesCondition() metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableDeliveryReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.MultipleMatchesDeliveryReadyReason,
		Message: "deliverable may only match a single delivery's selector",
	}
}

func MissingReadyInDeliveryCondition(deliveryReadyCondition metav1.Condition) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableDeliveryReady,
		Status:  metav1.ConditionFalse,
		Reason:  deliveryReadyCondition.Reason,
		Message: deliveryReadyCondition.Message,
	}
}

// -- Resource conditions

func ResourcesSubmittedCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.DeliverableResourcesSubmitted,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.CompleteResourcesSubmittedReason,
	}
}

func TemplateObjectRetrievalFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateObjectRetrievalFailureResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func MissingValueAtPathCondition(resourceName, expression string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionUnknown,
		Reason:  v1alpha1.MissingValueAtPathResourcesSubmittedReason,
		Message: fmt.Sprintf("Resource '%s' is waiting to read value '%s'", resourceName, expression),
	}
}

func TemplateStampFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func TemplateStampFailureByObservedGenerationCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureResourcesSubmittedReason,
		Message: fmt.Sprintf("resource '%s' cannot satisfy observedCompletion without observedGeneration in object status", err.(deliverable.RetrieveOutputError).ResourceName()),
	}
}

func DeploymentConditionNotMetCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionUnknown,
		Reason:  v1alpha1.DeploymentConditionNotMetResourcesSubmittedReason,
		Message: fmt.Sprintf("resource '%s' condition not met: %s", err.(deliverable.RetrieveOutputError).ResourceName(), err.(deliverable.RetrieveOutputError).Err.Error()),
	}
}

func DeploymentFailedConditionMetCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.DeploymentFailedConditionMetResourcesSubmittedReason,
		Message: fmt.Sprintf("resource '%s' failed condition met: %s", err.(deliverable.RetrieveOutputError).ResourceName(), err.(deliverable.RetrieveOutputError).Err.Error()),
	}
}

func TemplateRejectedByAPIServerCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateRejectedByAPIServerResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func UnknownResourceErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.DeliverableResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.UnknownErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}
