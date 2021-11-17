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
	"fmt"
)

type JsonPathError struct {
	Err        error
	expression string
}

func NewJsonPathError(expression string, err error) JsonPathError {
	return JsonPathError{
		Err:        err,
		expression: expression,
	}
}

func (e JsonPathError) Error() string {
	return fmt.Errorf("evaluate json path '%s': %w", e.expression, e.Err).Error()
}

func (e JsonPathError) JsonPathExpression() string {
	return e.expression
}

type ObservedGenerationError struct {
	Err error
}

func NewObservedGenerationError(err error) ObservedGenerationError {
	return ObservedGenerationError{
		Err: err,
	}
}

func (e ObservedGenerationError) Error() string {
	return e.Err.Error()
}

type DeploymentConditionError struct {
	Err error
}

func NewDeploymentConditionError(err error) DeploymentConditionError {
	return DeploymentConditionError{
		Err: err,
	}
}

func (e DeploymentConditionError) Error() string {
	return e.Err.Error()
}

type DeploymentFailedConditionMetError struct {
	Err error
}

func NewDeploymentFailedConditionMetError(err error) DeploymentFailedConditionMetError {
	return DeploymentFailedConditionMetError{
		Err: err,
	}
}

func (e DeploymentFailedConditionMetError) Error() string {
	return e.Err.Error()
}
