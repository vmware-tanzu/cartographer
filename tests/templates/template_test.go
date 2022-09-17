package templates

import (
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
		IgnoreMetadata:     false,
		IgnoreOwnerRefs:    true,
		IgnoreLabels:       true,
	}

	ts.Run(t)
}
