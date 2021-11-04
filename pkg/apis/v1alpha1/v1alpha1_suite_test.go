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

package v1alpha1_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

func TestV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V1alpha1 Suite")
}

// Test Helpers

type ArbitraryObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ArbitrarySpec
}

type ArbitrarySpec struct {
	SomeKey string `json:"someKey"`
}

func markersFor(relativeFile, typeName, fieldName, markerType string) (interface{}, error) {
	packages, err := loader.LoadRoots(relativeFile)
	if err != nil {
		return nil, err
	}
	if len(packages) != 1 {
		return nil, fmt.Errorf("got %d package(s) for file: %s", len(packages), relativeFile)
	}
	if len(packages[0].GoFiles) != 1 {
		return nil, fmt.Errorf("got %d GoFiles(s) for file: %s", len(packages[0].GoFiles), relativeFile)
	}

	// create a registry of CRD markers
	reg := &markers.Registry{}
	err = crdmarkers.Register(reg)
	if err != nil {
		return nil, err
	}

	// and a collector which `EachType` requires
	coll := &markers.Collector{Registry: reg}

	var mrkrs interface{}
	err = markers.EachType(coll, packages[0], func(info *markers.TypeInfo) {
		if info.Name == typeName {
			for _, fieldInfo := range info.Fields {
				if fieldInfo.Name == fieldName {
					mrkrs = fieldInfo.Markers.Get(markerType)
				}
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error enumerating types: %w", err)
	}

	if mrkrs == nil {
		return nil, fmt.Errorf(
			"could not find marker type '%s' in file/type/field: %s/%s/%s",
			markerType,
			relativeFile,
			typeName,
			fieldName,
		)
	}

	return mrkrs, nil
}
