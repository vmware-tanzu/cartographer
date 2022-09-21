package templates

import (
	"encoding/json"
	"fmt"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
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

	testSuite := template_tester.TemplateTestSuite{
		"template, workload and expected defined in files": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				WorkloadFile:    "workload.yaml",
				BlueprintParams: params,
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
			},
		},

		"template defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				Template:        templateOfDeliverable,
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
			},
			IgnoreMetadataFields: []string{"creationTimestamp"},
		},

		"expected defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				Template:        templateOfDeliverable,
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObject: expectedDeliverable,
			},
			IgnoreMetadata: true,
		},

		"workload defined as an object": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template.yaml",
				Workload:        workload,
				BlueprintParams: params,
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
			},
		},

		"clustertemplate uses ytt field": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile:    "template-ytt.yaml",
				BlueprintParams: params,
				WorkloadFile:    "workload.yaml",
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
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
				ExpectedObjectFile: "expected.yaml",
			},
		},

		"template requires ytt preprocessing, data supplied in files": {
			Inputs: template_tester.TemplateTestInputs{
				TemplateFile: "template-requires-preprocess.yaml",
				WorkloadFile: "workload.yaml",
				YttFiles:     []string{"values.yaml"},
			},
			Expectations: template_tester.TemplateTestExpectations{
				ExpectedObjectFile: "expected.yaml",
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

	url := "https://github.com/vmware-tanzu/cartographer/"
	branch := "main"

	deliverable.Spec.Source.Git.URL = &url
	deliverable.Spec.Source.Git.Ref.Branch = &branch
	newDeliverable.Spec.Params[0].Value = apiextensionsv1.JSON{Raw: []byte(`"some-secret"`)}
	newDeliverable.Spec.ServiceAccountName = "such-a-good-sa"

	return &newDeliverable
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
