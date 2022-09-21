package templates

import (
	"encoding/json"
	"fmt"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
)

func TestSupplyChainSourceTemplate(t *testing.T) {
	params, err := helpers.BuildBlueprintStringParams(helpers.StringParams{
		{
			Name:  "serviceAccount",
			Value: "my-sc",
		},
		{
			Name:         "gitImplementation",
			DefaultValue: "some-implementation",
		},
	})
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

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
		Spec: v1alpha1.WorkloadSpec{Source: &v1alpha1.Source{Git: &v1alpha1.GitSource{URL: &url, Ref: &v1alpha1.GitRef{Branch: &branch}}}},
	}

	testSuite := helpers.TemplateTestSuite{
		"workload as an object": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile:    "source.yaml",
				BlueprintParams: params,
				Workload:        &workload,
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
			},
		},

		"workload as a file": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile:    "source.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
			},
		},
	}

	testSuite.Run(t)
}

func TestAnother(t *testing.T) {
	params, err := helpers.BuildBlueprintStringParams(helpers.StringParams{
		{
			Name:         "gitops_url",
			DefaultValue: "https://github.com/waciumawanjohi/computer-science",
		},
		{
			Name:         "gitops_branch",
			DefaultValue: "main",
		},
	})
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

	deliverable := createDeliverable()

	templateOfDeliverable, err := createTemplate(deliverable)
	if err != nil {
		t.Fatalf("failed to create template: %v", err)
	}

	expectedDeliverable := createExpectedDeliverable(deliverable)

	testSuite := helpers.TemplateTestSuite{
		"Test params in regular template": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile:    "another-template-1.yaml",
				BlueprintParams: params,
				WorkloadFile:    "another-workload.yaml",
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "another-expect.yaml",
			},
		},

		"test params in ytt template": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile:    "another-template.yaml",
				BlueprintParams: params,
				WorkloadFile:    "another-workload.yaml",
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "another-expect.yaml",
			},
		},

		"template as an object": {
			Inputs: helpers.TemplateTestInputs{
				Template:        templateOfDeliverable,
				BlueprintParams: params,
				WorkloadFile:    "another-workload.yaml",
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "another-expect.yaml",
			},
			IgnoreMetadataFields: []string{"creationTimestamp"},
		},

		"expected as object": {
			Inputs: helpers.TemplateTestInputs{
				Template:        templateOfDeliverable,
				BlueprintParams: params,
				WorkloadFile:    "another-workload.yaml",
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObject: expectedDeliverable,
			},
			IgnoreMetadata: true,
		},

		"template modified by ytt": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile:    "another-template-preytt.yaml",
				BlueprintParams: params,
				WorkloadFile:    "another-workload.yaml",
				YttValues: helpers.Values{
					"kind": "Deliverable",
				},
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "another-expect.yaml",
			},
		},

		"template modified by ytt object": {
			Inputs: helpers.TemplateTestInputs{
				TemplateFile: "another-template-preytt.yaml",
				WorkloadFile: "another-workload.yaml",
				YttFiles:     []string{"values.yaml"},
			},
			Expectations: helpers.TemplateTestExpectations{
				ExpectedObjectFile: "another-expect.yaml",
			},
		},
	}

	testSuite.Run(t)
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

func createExpectedDeliverable(deliverable *v1alpha1.Deliverable) *v1alpha1.Deliverable {
	newDeliverable := *deliverable

	url := "https://github.com/waciumawanjohi/computer-science"
	branch := "main"

	deliverable.Spec.Source.Git.URL = &url
	deliverable.Spec.Source.Git.Ref.Branch = &branch
	newDeliverable.Spec.Params[0].Value = apiextensionsv1.JSON{Raw: []byte(`"some-secret"`)}
	newDeliverable.Spec.ServiceAccountName = "such-a-good-sa"

	return &newDeliverable
}
