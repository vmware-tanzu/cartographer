package templates

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSupplyChainSourceTemplate(t *testing.T) {
	params := map[string]interface{}{
		"serviceAccount":    "my-sc",
		"gitImplementation": "some-implementation",
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

	ts := helpers.TemplateTestSuite{
		TemplateFile:       "source.yaml",
		ExpectedObjectFile: "expected.yaml",
		Params:             params,
		Workload:           &workload,
	}

	ts.Run(t)

	ts = helpers.TemplateTestSuite{
		TemplateFile:       "source.yaml",
		ExpectedObjectFile: "expected.yaml",
		Params:             params,
		WorkloadFile:       "workload.yaml",
		Labels:             map[string]string{},
	}

	ts.Run(t)
}
