package conditions

import (
	"fmt"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -- Resource.Conditions - ResourcesHealthy - True

func OutputAvailableResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.OutputAvailableResourcesHealthyReason,
	}
}

func AlwaysHealthyResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: metav1.ConditionTrue,
		Reason: v1alpha1.AlwaysHealthyResourcesHealthyReason,
	}
}

func SingleConditionMatchCondition(status metav1.ConditionStatus, conditionName string) metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: status,
		Reason: fmt.Sprintf("%sCondition", conditionName),
	}
}

// -- Resource.Conditions - ResourcesHealthy - Unknown

func UnknownResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: metav1.ConditionUnknown,
	}
}

func SingleConditionTypeEvaluationErrorCondition(err error) metav1.Condition {
	return metav1.Condition{
		Type:    v1alpha1.Healthy,
		Status:  metav1.ConditionUnknown,
		Reason:  v1alpha1.SingleConditionTypeEvaluationErrorResourcesHealthyReason,
		Message: err.Error(),
	}
}

func SingleConditionTypeNoResultResourcesCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: metav1.ConditionUnknown,
		Reason: v1alpha1.SingleConditionTypeNoResultResourcesHealthyReason,
	}
}

// -- Resource.Conditions - ResourcesHealthy - False

func OutputNotAvailableResourcesHealthyCondition() metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: metav1.ConditionFalse,
		Reason: v1alpha1.OutputNotAvailableResourcesHealthyReason,
	}
}

func MultiMatchResourcesHealthyCondition(status metav1.ConditionStatus) metav1.Condition {
	return metav1.Condition{
		Type:   v1alpha1.Healthy,
		Status: status,
		Reason: v1alpha1.MultiMatchedResourcesHealthyReason,
	}
}
