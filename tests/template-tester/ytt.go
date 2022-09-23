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

package template_tester

import (
	"context"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type ytt struct {
	files  []string
	values []Values

	argv     []string
	tmpFiles []string
}

func YTT() *ytt {
	return &ytt{
		argv: []string{"ytt", "--ignore-unknown-comments"},
	}
}

func (y *ytt) Values(v Values) *ytt {
	y.values = append(y.values, v)
	return y
}

func (y *ytt) F(fpath string) *ytt {
	y.files = append(y.files, fpath)
	return y
}

func (y *ytt) build() error {
	for _, v := range y.values {
		f, err := os.CreateTemp("", "")
		if err != nil {
			return fmt.Errorf("create tmp: %w", err)
		}

		defer f.Close()

		if err := v.Encode(f); err != nil {
			return fmt.Errorf("values '%v' encode: %w", v, err)
		}

		y.argv = append(y.argv, "--data-values-file", f.Name())
		y.tmpFiles = append(y.tmpFiles, f.Name())
	}

	for _, fpath := range y.files {
		y.argv = append(y.argv, "-f", fpath)
	}

	return nil
}

func (y *ytt) gc() error {
	for _, fpath := range y.tmpFiles {
		if err := os.RemoveAll(fpath); err != nil {
			return fmt.Errorf("removeall '%s': %w", fpath, err)
		}
	}

	return nil
}

// ToTempFile runs `ytt` writing the results to a temporary file.
func (y *ytt) ToTempFile(ctx context.Context) (*os.File, error) {
	if err := y.build(); err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	// defer y.gc()

	f, err := os.CreateTemp("", "")
	if err != nil {
		return nil, fmt.Errorf("create temp: %w", err)
	}

	stdout, _, err := Cmd(y.argv...).RunWithOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("ytt: %w", err)
	}

	_, err = f.Write([]byte(stdout))
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return f, nil
}

type Values map[string]interface{}

func (v Values) Set(kvk string, kvv interface{}) {
	v[kvk] = kvv
}

func (v Values) Encode(w io.Writer) error {
	if err := yaml.NewEncoder(w).Encode(&v); err != nil {
		return fmt.Errorf("yaml encode: %w", err)
	}

	return nil
}
