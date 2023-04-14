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
)

// -- Workload.Status.Conditions - SupplyChainReady

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

func MissingPassThroughInputCondition(input string, resource string) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.WorkloadSupplyChainReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.PassThroughReason,
		Message: fmt.Sprintf("unable to find passthrough [%s] in [%s]", input, resource),
	}
}

func AddConditionForResourceSubmittedWorkload(conditionManager *ConditionManager, isOwner bool, err error) {
	switch typedErr := err.(type) {
	case cerrors.GetTemplateError:
		(*conditionManager).AddPositive(TemplateObjectRetrievalFailureCondition(isOwner, typedErr))
	case cerrors.StampError:
		(*conditionManager).AddPositive(TemplateStampFailureCondition(isOwner, typedErr))
	case cerrors.ApplyStampedObjectError:
		(*conditionManager).AddPositive(TemplateRejectedByAPIServerCondition(isOwner, typedErr))
	case cerrors.ListCreatedObjectsError:
		(*conditionManager).AddPositive(BlueprintsFailedToListCreatedObjectsCondition(isOwner, typedErr))
	case cerrors.NoHealthyImmutableObjectsError:
		(*conditionManager).AddPositive(NoHealthyImmutableObjectsCondition(isOwner, typedErr))
	case cerrors.RetrieveOutputError:
		if typedErr.StampedObject == nil {
			(*conditionManager).AddPositive(MissingPassThroughInputCondition(typedErr.PassThroughInput, typedErr.GetQualifiedResource()))
		} else {
			(*conditionManager).AddPositive(MissingValueAtPathCondition(isOwner, typedErr.StampedObject, typedErr.JsonPathExpression(), typedErr.GetQualifiedResource()))
		}
	case cerrors.ResolveTemplateOptionError:
		(*conditionManager).AddPositive(ResolveTemplateOptionsErrorCondition(isOwner, typedErr))
	case cerrors.TemplateOptionsMatchError:
		(*conditionManager).AddPositive(TemplateOptionsMatchErrorCondition(isOwner, typedErr))
	default:
		(*conditionManager).AddPositive(UnknownResourceErrorCondition(isOwner, typedErr))
	}
}
