package templates

import (
	"encoding/json"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
)

func TestSupplyChainSourceTemplate(t *testing.T) {
	param1, err := helpers.BuildBlueprintStringParam(
		"serviceAccount",
		"my-sc",
		"",
	)
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

	param2, err := helpers.BuildBlueprintStringParam(
		"gitImplementation",
		"",
		"some-implementation",
	)
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

	params := []v1alpha1.BlueprintParam{*param1, *param2}

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

	ts := helpers.TemplateTestSuite{
		TemplateFile:       "source.yaml",
		ExpectedObjectFile: "expected.yaml",
		BlueprintParams:    params,
		Workload:           &workload,
	}

	ts.Run(t)

	ts = helpers.TemplateTestSuite{
		TemplateFile:       "source.yaml",
		ExpectedObjectFile: "expected.yaml",
		BlueprintParams:    params,
		WorkloadFile:       "workload.yaml",
		Labels:             map[string]string{},
	}

	ts.Run(t)
}

func TestAnother(t *testing.T) {
	param1, err := helpers.BuildBlueprintStringParam(
		"gitops_url",
		"",
		"https://github.com/waciumawanjohi/computer-science",
	)
	param2, err := helpers.BuildBlueprintStringParam(
		"gitops_branch",
		"",
		"main",
	)
	if err != nil {
		t.Fatalf("failed to build param: %v", err)
	}

	ts := helpers.TemplateTestSuite{
		TemplateFile:       "another-template-1.yaml",
		ExpectedObjectFile: "another-expect.yaml",
		BlueprintParams:    []v1alpha1.BlueprintParam{*param1, *param2}, // TODO: simplify so users don't have to know about this internal struct
		WorkloadFile:       "another-workload.yaml",
	}

	ts.Run(t)

	ts = helpers.TemplateTestSuite{
		TemplateFile:       "another-template.yaml",
		ExpectedObjectFile: "another-expect.yaml",
		BlueprintParams:    []v1alpha1.BlueprintParam{*param1, *param2}, // TODO: simplify so users don't have to know about this internal struct
		WorkloadFile:       "another-workload.yaml",
	}

	ts.Run(t)

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

	dbytes, err := json.Marshal(deliverable)
	if err != nil {
		t.Fatalf("marshal deliverable: %v", err)
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

	ts = helpers.TemplateTestSuite{
		Template:             &template,
		ExpectedObjectFile:   "another-expect.yaml",
		BlueprintParams:      []v1alpha1.BlueprintParam{*param1, *param2},
		WorkloadFile:         "another-workload.yaml",
		IgnoreMetadataFields: []string{"creationTimestamp"},
	}

	ts.Run(t)

	deliverableURL = "https://github.com/waciumawanjohi/computer-science"
	deliverableBranch = "main"
	deliverable.Spec.Params[0].Value = apiextensionsv1.JSON{Raw: []byte(`"some-secret"`)}
	deliverable.Spec.ServiceAccountName = "such-a-good-sa"

	ts = helpers.TemplateTestSuite{
		Template:        &template,
		ExpectedObject:  deliverable,
		BlueprintParams: []v1alpha1.BlueprintParam{*param1, *param2},
		WorkloadFile:    "another-workload.yaml",
		IgnoreMetadata:  true,
	}

	ts.Run(t)
}
