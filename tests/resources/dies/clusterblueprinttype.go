package dies

import (
	"carto.run/blueprints/api/v1alpha1"
)

// +die:object=true
type _ = v1alpha1.ClusterBlueprintType

// +die
type _ = v1alpha1.ClusterBlueprintTypeSpec

// +die
type _ = v1alpha1.ClusterBlueprintTypeStatus
