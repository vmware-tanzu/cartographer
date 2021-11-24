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

package runnable

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

// -- ClusterRunTemplate conditions

func RunTemplateReadyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.RunTemplateReady,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.ReadyRunTemplateReason,
	}
}

func RunTemplateMissingCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.NotFoundRunTemplateReason,
		Message: err.Error(),
	}
}

func StampedObjectRejectedByAPIServerCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.StampedObjectRejectedByAPIServerRunTemplateReason,
		Message: err.Error(),
	}
}

func OutputPathNotSatisfiedCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.OutputPathNotSatisfiedRunTemplateReason,
		Message: err.Error(),
	}
}

func FailedToListCreatedObjectsCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.FailedToListCreatedObjectsReason,
		Message: err.Error(),
	}
}

func TemplateStampFailureCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.TemplateStampFailureRunTemplateReason,
		Message: err.Error(),
	}
}

func UnknownErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.UnknownErrorReason,
		Message: err.Error(),
	}
}

func ServiceAccountSecretNotFoundCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ServiceAccountSecretErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}

func ClientBuilderErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.RunTemplateReady,
		Status:  metav1.ConditionFalse,
		Reason:  v1alpha1.ClientBuilderErrorResourcesSubmittedReason,
		Message: err.Error(),
	}
}
