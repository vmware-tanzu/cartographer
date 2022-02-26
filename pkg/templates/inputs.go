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

type SourceInput struct {
	URL      string `json:"url"`
	Revision string `json:"revision"`
	Name     string      `json:"name"`
}

type ImageInput struct {
	Image string `json:"image"`
	Name  string      `json:"name"`
}

type ConfigInput struct {
	Config string `json:"config"`
	Name   string      `json:"name"`
}

type Inputs struct {
	Sources    map[string]SourceInput
	Images     map[string]ImageInput
	Configs    map[string]ConfigInput
	Deployment *SourceInput
}

func (i Inputs) OnlySource() *SourceInput {
	if len(i.Sources) == 1 {
		for _, sourceInput := range i.Sources {
			return &sourceInput
		}
	}
	return nil
}

func (i Inputs) OnlyImage() string {
	if len(i.Images) == 1 {
		for _, imageInput := range i.Images {
			return imageInput.Image
		}
	}
	return ""
}

func (i Inputs) OnlyConfig() string {
	if len(i.Configs) == 1 {
		for _, configInput := range i.Configs {
			return configInput.Config
		}
	}
	return ""
}
