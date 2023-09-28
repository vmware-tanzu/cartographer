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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	cerrors "github.com/vmware-tanzu/cartographer/pkg/errors"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
)

// -- Deliverable.Status.Conditions - DeliveryReady

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

// -- Deliverable.Status.Resource[x].Conditions - ResourceSubmitted &&
// -- Deliverable.Status.Conditions - ResourcesSubmitted

func TemplateStampFailureByObservedGenerationCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.OwnerResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureResourcesSubmittedReason,
		Message: fmt.Sprintf("resource [%s] cannot satisfy observedCompletion without observedGeneration in object status", err.(cerrors.RetrieveOutputError).GetResourceName()),
	}
}

func DeploymentConditionNotMetCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.OwnerResourcesSubmitted,
		Status:  metav1.ConditionUnknown,
		Reason:  v1alpha1.DeploymentConditionNotMetResourcesSubmittedReason,
		Message: fmt.Sprintf("resource [%s] condition not met: %s", err.(cerrors.RetrieveOutputError).GetResourceName(), err.(cerrors.RetrieveOutputError).Err.Error()),
	}
}

func DeploymentFailedConditionMetCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.OwnerResourcesSubmitted,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.DeploymentFailedConditionMetResourcesSubmittedReason,
		Message: fmt.Sprintf("resource [%s] failed condition met: %s", err.(cerrors.RetrieveOutputError).GetResourceName(), err.(cerrors.RetrieveOutputError).Err.Error()),
	}
}

func AddConditionForResourceSubmittedDeliverable(conditionManager *ConditionManager, isOwner bool, err error) {
	switch typedErr := err.(type) {
	case cerrors.GetTemplateError:
		(*conditionManager).AddPositive(TemplateObjectRetrievalFailureCondition(isOwner, typedErr))
	case cerrors.StampError:
		(*conditionManager).AddPositive(TemplateStampFailureCondition(isOwner, typedErr))
	case cerrors.ApplyStampedObjectError:
		(*conditionManager).AddPositive(TemplateRejectedByAPIServerCondition(isOwner, typedErr))
	case cerrors.RetrieveOutputError:
		switch typedErr.Err.(type) {
		case stamp.ObservedGenerationError:
			(*conditionManager).AddPositive(TemplateStampFailureByObservedGenerationCondition(typedErr))
		case stamp.DeploymentFailedConditionMetError:
			(*conditionManager).AddPositive(DeploymentFailedConditionMetCondition(typedErr))
		case stamp.DeploymentConditionError:
			(*conditionManager).AddPositive(DeploymentConditionNotMetCondition(typedErr))
		case stamp.JsonPathError:
			(*conditionManager).AddPositive(MissingValueAtPathCondition(isOwner, typedErr.StampedObject, typedErr.JsonPathExpression(), typedErr.GetQualifiedResource(), typedErr.Healthy))
		default:
			(*conditionManager).AddPositive(UnknownResourceErrorCondition(isOwner, typedErr))
		}
	case cerrors.ResolveTemplateOptionError:
		(*conditionManager).AddPositive(ResolveTemplateOptionsErrorCondition(isOwner, typedErr))
	case cerrors.TemplateOptionsMatchError:
		(*conditionManager).AddPositive(TemplateOptionsMatchErrorCondition(isOwner, typedErr))
	default:
		(*conditionManager).AddPositive(UnknownResourceErrorCondition(isOwner, typedErr))
	}
}
