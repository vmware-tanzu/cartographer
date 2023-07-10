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

import (
	"testing"

	cartotesting "github.com/vmware-tanzu/cartographer/pkg/testing"
)

func TestCLIExample(t *testing.T) {
	directories := []string{"kpack", "deliverable", "deployment"}

	for _, directory := range directories {
		err := cartotesting.CliTest(directory)
		if err != nil {
			t.Fatal("cli test failed,", "directory:", directory, "err:", err)
		}
	}
}
