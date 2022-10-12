package controllers

import (
	"context"

	"github.com/vmware-tanzu/cartographer/pkg/realizer"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/statuses"
)

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Realizer
type Realizer interface {
	Realize(ctx context.Context, resourceRealizer realizer.ResourceRealizer, blueprintName string, ownerResources []realizer.OwnerResource, resourceStatuses statuses.ResourceStatuses) error
}
