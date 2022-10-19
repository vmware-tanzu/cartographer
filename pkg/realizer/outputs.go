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

package realizer

import (
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

func NewOutputs() Outputs {
	return make(Outputs)
}

type Outputs map[string]*templates.Output

func (o Outputs) AddOutput(resourceName string, output *templates.Output) {
	o[resourceName] = output
}

func (o Outputs) GetImage(resourceName string) templates.Image {
	output := o[resourceName]
	if output == nil {
		return nil
	}
	return output.Image
}

func (o Outputs) GetConfig(resourceName string) templates.Config {
	output := o[resourceName]
	if output == nil {
		return nil
	}
	return output.Config
}

func (o Outputs) GetSource(resourceName string) *templates.Source {
	output := o[resourceName]
	if output == nil {
		return nil
	}

	return output.Source
}
