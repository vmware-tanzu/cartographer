package resources

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceStatuses interface {
	GetPreviousRealizedResource(name string) *v1alpha1.RealizedResource
	Add(status *v1alpha1.RealizedResource, err error)
	GetCurrent() []v1alpha1.ResourceStatus
	IsChanged() bool
}

func NewResourceStatuses(previousResourceStatuses []v1alpha1.ResourceStatus) *resourceStatuses {
	statuses := map[string]*resourceStatus{}

	for _, previousResourceStatus := range previousResourceStatuses {
		statuses[previousResourceStatus.Name] = &resourceStatus{
			Previous: &v1alpha1.ResourceStatus{
				RealizedResource: previousResourceStatus.RealizedResource,
			},
		}
	}

	return &resourceStatuses{
		statuses: statuses,
	}
}

type resourceStatus struct {
	Previous *v1alpha1.ResourceStatus
	Current  *v1alpha1.ResourceStatus
}

type resourceStatuses struct {
	statuses map[string]*resourceStatus
}

func (r resourceStatuses) IsChanged() bool {
	// TODO
	return false
}

func (r resourceStatuses) GetCurrent() []v1alpha1.ResourceStatus {
	var currentStatuses []v1alpha1.ResourceStatus

	// TODO: if current is nil, we need to use previous?
	// ...In the case of reconciler failing before we get to the realizer
	for _, status := range r.statuses {
		currentStatuses = append(currentStatuses, *status.Current)
	}

	return currentStatuses
}

func (r resourceStatuses) GetPreviousRealizedResource(name string) *v1alpha1.RealizedResource {
	if r.statuses[name] != nil && &r.statuses[name].Previous != nil {
		return &r.statuses[name].Previous.RealizedResource
	}
	return nil
}

func (r resourceStatuses) Add(realizedResource *v1alpha1.RealizedResource, err error) {
	name := realizedResource.Name

	if r.statuses[name] == nil {
		r.statuses[name] = &resourceStatus{}
	}

	r.statuses[name] = &resourceStatus{
		Previous: nil,
		Current: &v1alpha1.ResourceStatus{
			RealizedResource: *realizedResource,
			Conditions:       r.createConditions(name, err),
		},
	}
}

func (r resourceStatuses) createConditions(name string, err error) []metav1.Condition {
	var previousConditions []metav1.Condition
	if r.statuses[name].Previous != nil {
		previousConditions = r.statuses[name].Previous.Conditions
	}

	conditionManager := conditions.NewConditionManager(v1alpha1.ResourceReady, previousConditions)
	if err != nil {
		conditions.AddConditionForResourceSubmitted(&conditionManager, false, err)
	} else {
		conditionManager.AddPositive(conditions.ResourceSubmittedCondition())
	}

	resourceConditions, _ := conditionManager.Finalize()
	return resourceConditions
}
