package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SchemaValid is a condition indicating that the ClusterBlueprintTypeSpec.Schema is valid.
// Valid Reasons are:
//   SchemaInvalidReason
//   SchemaValidReason
type SchemaValid struct {
	metav1.Condition
}

// SchemaValidReason indicates that ClusterBlueprintTypeSpec.Schema is valid
func SchemaValidReason() SchemaValid {
	return SchemaValid{
		Condition: metav1.Condition{
			Type:   "SchemaValid",
			Status: metav1.ConditionTrue,
		},
	}
}

// SchemaInvalidReason describes a syntax error in the provided ClusterBlueprintTypeSpec.Schema
func SchemaInvalidReason(message string) SchemaValid {
	return SchemaValid{
		Condition: metav1.Condition{
			Type:    "SchemaInvalid",
			Status:  metav1.ConditionFalse,
			Reason:  "MalformedOpenAPISchema",
			Message: message,
		},
	}
}
