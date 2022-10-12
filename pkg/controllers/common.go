package controllers

import (
	"context"

	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
)

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer realizer.ResourceRealizer, blueprintName string, ownerResources []realizer.OwnerResource, resourceStatuses statuses.ResourceStatuses) error
}
