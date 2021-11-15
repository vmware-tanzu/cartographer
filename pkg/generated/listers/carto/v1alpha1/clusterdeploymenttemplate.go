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
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/vmware-tanzu/cartographer/pkg/apis/carto/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ClusterDeploymentTemplateLister helps list ClusterDeploymentTemplates.
// All objects returned here must be treated as read-only.
type ClusterDeploymentTemplateLister interface {
	// List lists all ClusterDeploymentTemplates in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterDeploymentTemplate, err error)
	// ClusterDeploymentTemplates returns an object that can list and get ClusterDeploymentTemplates.
	ClusterDeploymentTemplates(namespace string) ClusterDeploymentTemplateNamespaceLister
	ClusterDeploymentTemplateListerExpansion
}

// clusterDeploymentTemplateLister implements the ClusterDeploymentTemplateLister interface.
type clusterDeploymentTemplateLister struct {
	indexer cache.Indexer
}

// NewClusterDeploymentTemplateLister returns a new ClusterDeploymentTemplateLister.
func NewClusterDeploymentTemplateLister(indexer cache.Indexer) ClusterDeploymentTemplateLister {
	return &clusterDeploymentTemplateLister{indexer: indexer}
}

// List lists all ClusterDeploymentTemplates in the indexer.
func (s *clusterDeploymentTemplateLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterDeploymentTemplate, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterDeploymentTemplate))
	})
	return ret, err
}

// ClusterDeploymentTemplates returns an object that can list and get ClusterDeploymentTemplates.
func (s *clusterDeploymentTemplateLister) ClusterDeploymentTemplates(namespace string) ClusterDeploymentTemplateNamespaceLister {
	return clusterDeploymentTemplateNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ClusterDeploymentTemplateNamespaceLister helps list and get ClusterDeploymentTemplates.
// All objects returned here must be treated as read-only.
type ClusterDeploymentTemplateNamespaceLister interface {
	// List lists all ClusterDeploymentTemplates in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ClusterDeploymentTemplate, err error)
	// Get retrieves the ClusterDeploymentTemplate from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ClusterDeploymentTemplate, error)
	ClusterDeploymentTemplateNamespaceListerExpansion
}

// clusterDeploymentTemplateNamespaceLister implements the ClusterDeploymentTemplateNamespaceLister
// interface.
type clusterDeploymentTemplateNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ClusterDeploymentTemplates in the indexer for a given namespace.
func (s clusterDeploymentTemplateNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ClusterDeploymentTemplate, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ClusterDeploymentTemplate))
	})
	return ret, err
}

// Get retrieves the ClusterDeploymentTemplate from the indexer for a given namespace and name.
func (s clusterDeploymentTemplateNamespaceLister) Get(name string) (*v1alpha1.ClusterDeploymentTemplate, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("clusterdeploymenttemplate"), name)
	}
	return obj.(*v1alpha1.ClusterDeploymentTemplate), nil
}
