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

package conditions

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate . ArtifactManager

// ArtifactManager supports collecting artifact statuses for your controller
// It adds a complete top level condition when Finalize is called.
type ArtifactManager interface {
	// Add a condition and associate a polarity with it.
	Add(artifact v1alpha1.Artifact)

	// Finalize	returns all artifacts
	// The changed result represents whether the conditions have changed enough to warrant an update to the APIServer
	Finalize() (artifacts []v1alpha1.Artifact, changed bool)
}

type artifactManager struct {
	previousArtifacts []v1alpha1.Artifact
	artifacts         []v1alpha1.Artifact
	changed           bool
}

type ArtifactManagerBuilder func(previousArtifacts []v1alpha1.Artifact) ArtifactManager

func NewArtifactManager(previousArtifacts []v1alpha1.Artifact) ArtifactManager {
	return &artifactManager{
		previousArtifacts: previousArtifacts,
	}
}

func create()

func (a *artifactManager) Add(artifact v1alpha1.Artifact) {
	isNewArtifact := true
	var previousArtifact v1alpha1.SubArtifact

	for _, previousArtifact = range a.previousArtifacts {
		if previousArtifact.GetID() == condition.Type {
			isNewArtifact = false
			lastTransitionTime := condition.LastTransitionTime
			condition.LastTransitionTime = previousArtifact.LastTransitionTime
			if !reflect.DeepEqual(previousArtifact, condition) {
				condition.LastTransitionTime = lastTransitionTime
				c.changed = true
			}
		}
	}

	if isNewArtifact {
		c.changed = true
	}

	c.conditions = append(c.conditions, condition)
}

func (a *artifactManager) Finalize() ([]v1alpha1.Artifact, bool) {
	return a.artifacts, a.changed
}
