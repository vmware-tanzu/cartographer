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

// TODO: This does not belong here, not sure where?

type SourceInput struct {
	URL      interface{} `json:"url"`
	Revision interface{} `json:"revision"`
	Name     string      `json:"name"`
}

type ImageInput struct {
	Image interface{} `json:"image"`
	Name  string      `json:"name"`
}

type ConfigInput struct {
	Config interface{} `json:"config"`
	Name   string      `json:"name"`
}
