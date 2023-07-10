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

package testing

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
)

type Workload interface {
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
