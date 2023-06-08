package testing

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type TTWorkload interface {
	GetWorkload() (*v1alpha1.Workload, error)
}

type WorkloadObject struct {
	Workload *v1alpha1.Workload
}

func (w *WorkloadObject) GetWorkload() (*v1alpha1.Workload, error) {
	return w.Workload, nil
}

type WorkloadFile struct {
	Path string
}

func (w *WorkloadFile) GetWorkload() (*v1alpha1.Workload, error) {
	workload := &v1alpha1.Workload{}

	workloadData, err := os.ReadFile(w.Path)
	if err != nil {
		return nil, fmt.Errorf("could not read workload file: %w", err)
	}

	if err = yaml.Unmarshal(workloadData, workload); err != nil {
		return nil, fmt.Errorf("unmarshall template: %w", err)
	}

	return workload, nil
}
