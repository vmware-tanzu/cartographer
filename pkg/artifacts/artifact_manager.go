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

package artifacts

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/templates"
)

//counterfeiter:generate . ArtifactManager

// ArtifactManager supports collecting artifact statuses for your controller
// It adds a complete top level condition when Finalize is called.
type ArtifactManager interface {
	// Add a condition and associate a polarity with it.
	Add(artifact *v1alpha1.Artifact)

	// Finalize	returns all artifacts
	// The changed result represents whether the conditions have changed enough to warrant an update to the APIServer
	Finalize() (artifacts []v1alpha1.Artifact, changed bool)
}

type artifactManager struct {
	previousArtifacts []v1alpha1.Artifact
	artifacts         []v1alpha1.Artifact
	changed           bool
}

func NewArtifactManager(previousArtifacts []v1alpha1.Artifact) ArtifactManager {
	return &artifactManager{
		previousArtifacts: previousArtifacts,
	}
}

func CreateArtifact(passedContext templates.Stamper, resource *unstructured.Unstructured, output *templates.Output, resourceName string) (*v1alpha1.Artifact, error) {
	artifact := v1alpha1.Artifact{}
	artifact.Passed = v1alpha1.PassedResource{
		ResourceName:    resourceName,
		Kind:            resource.GetKind(),
		ApiVersion:      resource.GetAPIVersion(),
		Name:            resource.GetName(),
		Namespace:       resource.GetNamespace(),
		ResourceVersion: resource.GetResourceVersion(),
	}

	if output.Config != "" {
		artifact.Config.Config = string(output.Config)
	}

	if output.Image != "" {
		artifact.Image.Image = string(output.Image)
	}

	if output.Source != nil {
		artifact.Source.Revision = output.Source.Revision
		artifact.Source.Url = output.Source.URL
	}

	artifactJson, err := json.Marshal(artifact)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(artifactJson)

	artifact.Id = hex.EncodeToString(hash[:])

	return &artifact, nil
}

func (a *artifactManager) Add(artifact *v1alpha1.Artifact) {
	isNewArtifact := true

	for _, previousArtifact := range a.previousArtifacts {
		if previousArtifact.Id == artifact.Id {
			isNewArtifact = false
		}
	}

	if isNewArtifact {
		a.changed = true
	}

	a.artifacts = append(a.artifacts, *artifact)
}

func (a *artifactManager) Finalize() ([]v1alpha1.Artifact, bool) {
	return a.artifacts, a.changed
}
