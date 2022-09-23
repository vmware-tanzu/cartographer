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

package templates

import (
	"encoding/json"
	"fmt"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
	"github.com/vmware-tanzu/cartographer/tests/template-tester"
)

func TestTemplateExample(t *testing.T) {
	params, err := template_tester.BuildBlueprintStringParams(template_tester.StringParams{
		{
			Name:         "gitops_url",
			DefaultValue: "https://github.com/vmware-tanzu/cartographer/",
		},
		{
			Name:         "gitops_branch",
			DefaultValue: "main",
		},
	})
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

	workload := createWorkload()

	deliverable := createDeliverable()

	templateOfDeliverable, err := createTemplate(deliverable)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	expectedDeliverable := createExpectedDeliverable(deliverable)

	expectedUnstructured := createExpectedUnstructured()

	testSuite := template_tester.TemplateTestSuite{
		"template, workload and expected defined in files": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				WorkloadFile:    "workload.yaml",
				BlueprintParams: params,
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
		},

		"template defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				Template:        templateOfDeliverable,
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
			IgnoreMetadataFields: []string{"creationTimestamp"},
		},

		"workload defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				Workload:        workload,
				BlueprintParams: params,
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
		},

		"expected defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObject: expectedDeliverable,
			},
			IgnoreMetadata: true,
		},

		"expected defined as an unstructured": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedUnstructured: &expectedUnstructured,
			},
		},

		"clustertemplate uses ytt field": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template-ytt.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
		},

		"template requires ytt preprocessing, data supplied in object": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template-requires-preprocess.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
				YttValues: template_tester.Values{
					"kind": "Deliverable",
				},
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
		},

		"template requires ytt preprocessing, data supplied in files": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile: "template-requires-preprocess.yaml",
				WorkloadFile: "workload.yaml",
				YttFiles:     []string{"values.yaml"},
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected.yaml",
			},
		},

		"template that requires a supply chain input": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile: "template-kpack.yaml",
				WorkloadFile: "workload.yaml",
				SupplyChainInputs: templates.Inputs{
					Sources: map[string]templates.SourceInput{
						"source": {
							URL: "some-passed-on-url",
						},
					},
				},
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedFile: "expected-kpack.yaml",
			},
			IgnoreMetadata: true,
		},
	}

	testSuite.Run(t)
}

func createWorkload() *v1alpha1.Workload {
	url := "some-url"
	branch := "some-branch"

	workload := v1alpha1.Workload{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Workload",
			APIVersion: "carto.run/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
			Name:       "my-workload-name",
			Namespace:  "my-namespace",
		},
		Spec: v1alpha1.WorkloadSpec{
			ServiceAccountName: "such-a-good-sa",
			Params: []v1alpha1.OwnerParam{
				{
					Name:  "gitops_url",
					Value: apiextensionsv1.JSON{Raw: []byte(`"https://github.com/vmware-tanzu/cartographer/"`)},
				},
			},
			Source: &v1alpha1.Source{
				Git: &v1alpha1.GitSource{
					URL: &url,
					Ref: &v1alpha1.GitRef{
						Branch: &branch}}}},
	}

	return &workload
}

func createDeliverable() *v1alpha1.Deliverable {
	deliverableURL := `$(params.gitops_url)$`
	deliverableBranch := `$(params.gitops_branch)$`

	deliverable := &v1alpha1.Deliverable{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deliverable",
			APIVersion: "carto.run/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: `$(workload.metadata.name)$`,
		},
		Spec: v1alpha1.DeliverableSpec{
			ServiceAccountName: `$(workload.spec.serviceAccountName)$`,
			Params: []v1alpha1.OwnerParam{
				{
					Name:  "gitops_ssh_secret",
					Value: apiextensionsv1.JSON{Raw: []byte(`"$(params.gitops_ssh_secret)$"`)},
				},
			},
			Source: &v1alpha1.Source{
				Git: &v1alpha1.GitSource{
					URL: &deliverableURL,
					Ref: &v1alpha1.GitRef{
						Branch: &deliverableBranch,
					},
				},
			},
		},
	}
	return deliverable
}

func createTemplate(deliverable *v1alpha1.Deliverable) (*v1alpha1.ClusterTemplate, error) {
	dbytes, err := json.Marshal(deliverable)
	if err != nil {
		return nil, fmt.Errorf("marshal deliverable: %w", err)
	}

	template := v1alpha1.ClusterTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterTemplate",
			APIVersion: "carto.run/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "create-deliverable",
		},
		Spec: v1alpha1.TemplateSpec{
			Template: &runtime.RawExtension{Raw: dbytes},
			Params: []v1alpha1.TemplateParam{
				{
					Name:         "gitops_ssh_secret",
					DefaultValue: apiextensionsv1.JSON{Raw: []byte(`"some-secret"`)},
				},
			},
		},
	}
	return &template, nil
}

func createExpectedDeliverable(deliverable *v1alpha1.Deliverable) *v1alpha1.Deliverable {
	newDeliverable := *deliverable

	url := "https://github.com/vmware-tanzu/cartographer/"
	branch := "main"

	deliverable.Spec.Source.Git.URL = &url
	deliverable.Spec.Source.Git.Ref.Branch = &branch
	newDeliverable.Spec.Params[0].Value = apiextensionsv1.JSON{Raw: []byte(`"some-secret"`)}
	newDeliverable.Spec.ServiceAccountName = "such-a-good-sa"

	return &newDeliverable
}

func createExpectedUnstructured() unstructured.Unstructured {
	expectedUnstructured := unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "carto.run/v1alpha1",
			"kind":       "Deliverable",
			"metadata": map[string]any{
				"labels": map[string]any{
					"carto.run/cluster-template-name": "create-deliverable",
					"carto.run/template-kind":         "ClusterTemplate",
					"carto.run/workload-name":         "my-workload-name",
					"carto.run/workload-namespace":    "my-namespace",
				},
				"name":      "my-workload-name",
				"namespace": "my-namespace",
				"ownerReferences": []any{
					map[string]any{
						"apiVersion":         "carto.run/v1alpha1",
						"blockOwnerDeletion": true,
						"controller":         true,
						"kind":               "Workload",
						"name":               "my-workload-name",
						"uid":                "",
					},
				},
			},
			"spec": map[string]any{
				"params": []any{
					map[string]any{"name": "gitops_ssh_secret", "value": "some-secret"},
				},
				"serviceAccountName": "such-a-good-sa",
				"source": map[string]any{
					"git": map[string]any{
						"ref": map[string]any{
							"branch": "main",
						},
						"url": "https://github.com/vmware-tanzu/cartographer/",
					},
				},
			},
		},
	}
	return expectedUnstructured
}
