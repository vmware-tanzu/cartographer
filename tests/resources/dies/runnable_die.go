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

package dies

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// +die:object=true,spec=RunnableSpec,status=RunnableStatus
type _ = v1alpha1.Runnable

// +die:ignore={Inputs}
type _ = v1alpha1.RunnableSpec

func (d *RunnableSpecDie) SetInputString(key string, value string) *RunnableSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunnableSpec) {
		if r.Inputs == nil {
			r.Inputs = map[string]apiextensionsv1.JSON{}
		}

		r.Inputs[key] = apiextensionsv1.JSON{Raw: []byte(`"` + value + `"`)}
	})
}

// +die
type _ = v1alpha1.RunnableStatus
