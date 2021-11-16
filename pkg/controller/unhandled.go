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

package controller

import (
	"errors"
)

func NewUnhandledError(err error) error {
	return unhandledError{e: err}
}

type unhandledError struct {
	e error
}

func IsUnhandledError(err error) bool {
	return errors.Is(err, unhandledError{})
}

func (unhandledError) Is(target error) bool {
	//nolint: errorlint // This check is actually fine.
	_, ok := target.(unhandledError)
	return ok
}

func (err unhandledError) Error() string {
	if err.e == nil {
		return ""
	}

	return err.e.Error()
}

func (err unhandledError) Unwrap() error {
	return err.e
}
