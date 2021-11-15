/*
Copyright 2021 VMware

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/vmware-tanzu/cartographer/pkg/apis/carto/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterConfigTemplates implements ClusterConfigTemplateInterface
type FakeClusterConfigTemplates struct {
	Fake *FakeCartoV1alpha1
	ns   string
}

var clusterconfigtemplatesResource = schema.GroupVersionResource{Group: "carto.run", Version: "v1alpha1", Resource: "clusterconfigtemplates"}

var clusterconfigtemplatesKind = schema.GroupVersionKind{Group: "carto.run", Version: "v1alpha1", Kind: "ClusterConfigTemplate"}

// Get takes name of the clusterConfigTemplate, and returns the corresponding clusterConfigTemplate object, and an error if there is any.
func (c *FakeClusterConfigTemplates) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ClusterConfigTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clusterconfigtemplatesResource, c.ns, name), &v1alpha1.ClusterConfigTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterConfigTemplate), err
}

// List takes label and field selectors, and returns the list of ClusterConfigTemplates that match those selectors.
func (c *FakeClusterConfigTemplates) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ClusterConfigTemplateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clusterconfigtemplatesResource, clusterconfigtemplatesKind, c.ns, opts), &v1alpha1.ClusterConfigTemplateList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ClusterConfigTemplateList{ListMeta: obj.(*v1alpha1.ClusterConfigTemplateList).ListMeta}
	for _, item := range obj.(*v1alpha1.ClusterConfigTemplateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterConfigTemplates.
func (c *FakeClusterConfigTemplates) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clusterconfigtemplatesResource, c.ns, opts))

}

// Create takes the representation of a clusterConfigTemplate and creates it.  Returns the server's representation of the clusterConfigTemplate, and an error, if there is any.
func (c *FakeClusterConfigTemplates) Create(ctx context.Context, clusterConfigTemplate *v1alpha1.ClusterConfigTemplate, opts v1.CreateOptions) (result *v1alpha1.ClusterConfigTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clusterconfigtemplatesResource, c.ns, clusterConfigTemplate), &v1alpha1.ClusterConfigTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterConfigTemplate), err
}

// Update takes the representation of a clusterConfigTemplate and updates it. Returns the server's representation of the clusterConfigTemplate, and an error, if there is any.
func (c *FakeClusterConfigTemplates) Update(ctx context.Context, clusterConfigTemplate *v1alpha1.ClusterConfigTemplate, opts v1.UpdateOptions) (result *v1alpha1.ClusterConfigTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clusterconfigtemplatesResource, c.ns, clusterConfigTemplate), &v1alpha1.ClusterConfigTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterConfigTemplate), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClusterConfigTemplates) UpdateStatus(ctx context.Context, clusterConfigTemplate *v1alpha1.ClusterConfigTemplate, opts v1.UpdateOptions) (*v1alpha1.ClusterConfigTemplate, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(clusterconfigtemplatesResource, "status", c.ns, clusterConfigTemplate), &v1alpha1.ClusterConfigTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterConfigTemplate), err
}

// Delete takes name of the clusterConfigTemplate and deletes it. Returns an error if one occurs.
func (c *FakeClusterConfigTemplates) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(clusterconfigtemplatesResource, c.ns, name), &v1alpha1.ClusterConfigTemplate{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterConfigTemplates) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clusterconfigtemplatesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ClusterConfigTemplateList{})
	return err
}

// Patch applies the patch and returns the patched clusterConfigTemplate.
func (c *FakeClusterConfigTemplates) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ClusterConfigTemplate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterconfigtemplatesResource, c.ns, name, pt, data, subresources...), &v1alpha1.ClusterConfigTemplate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.ClusterConfigTemplate), err
}
