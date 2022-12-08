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

package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

const DefaultResyncTime = 10 * time.Hour

func GetObjectGVK(obj metav1.Object, scheme *runtime.Scheme) (schema.GroupVersionKind, error) {
	ro, ok := obj.(runtime.Object)
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("%T is not a runtime.Object", obj)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("gvk for object: %w", err)
	}

	return gvk, nil
}

func HereYaml(y string) string {
	y = strings.ReplaceAll(y, "\t", "    ")
	return heredoc.Doc(y)
}

func HereYamlF(y string, args ...interface{}) string {
	y = strings.ReplaceAll(y, "\t", "    ")
	return heredoc.Docf(y, args...)
}

func CreateNamespacedObjectOnClusterFromYamlDefinition(ctx context.Context, c client.Client, objYaml, namespace string) *unstructured.Unstructured {
	obj := createUnstructuredObject(objYaml)

	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	createObjectOnClusterFromUnstructured(ctx, c, obj)
	return obj
}

func CreateObjectOnClusterFromYamlDefinition(ctx context.Context, c client.Client, objYaml string) *unstructured.Unstructured {
	obj := createUnstructuredObject(objYaml)
	createObjectOnClusterFromUnstructured(ctx, c, obj)
	return obj
}

func createObjectOnClusterFromUnstructured(ctx context.Context, c client.Client, obj *unstructured.Unstructured) {
	err := c.Create(ctx, obj, &client.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func createUnstructuredObject(objYaml string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(objYaml), obj)
	Expect(err).NotTo(HaveOccurred())
	return obj
}

func UpdateObjectOnClusterFromYamlDefinition(ctx context.Context, c client.Client, newObjYaml string, originalObjNamespace string, origObjType client.Object) {
	newObj := createUnstructuredObject(newObjYaml)
	if originalObjNamespace != "" {
		newObj.SetNamespace(originalObjNamespace)
	}
	UpdateObjectOnCluster(ctx, c, newObj, origObjType)
}

func UpdateObjectOnCluster(ctx context.Context, c client.Client, newObj, origObjType client.Object) {
	Eventually(func() error {
		err := c.Get(ctx, client.ObjectKey{Name: newObj.GetName(), Namespace: newObj.GetNamespace()}, origObjType)
		if err != nil {
			return err
		}
		return c.Patch(ctx, newObj, client.MergeFromWithOptions(origObjType, client.MergeFromWithOptimisticLock{}))
	}).Should(Succeed())
}

func AlterFieldOfNestedStringMaps(obj interface{}, key string, value string) error {
	switch knownType := obj.(type) {
	case map[string]interface{}:
		i := strings.Index(key, ".")
		if i < 0 {
			_, ok := knownType[key]
			if !ok {
				return errors.New("field not found")
			}
			knownType[key] = value
		} else {
			keyPrefix := key[:i]
			keySuffix := key[i+1:]
			subMap, ok := knownType[keyPrefix]
			if !ok {
				return errors.New("field not found")
			}
			return AlterFieldOfNestedStringMaps(subMap, keySuffix, value)
		}

		return nil
	case []interface{}:
		if key[:1] != "[" {
			return errors.New("field not found")
		}
		closeBracket := strings.Index(key, "]")
		if closeBracket < 0 {
			return errors.New("field not found")
		}
		strIndex := key[1:closeBracket]
		intIndex, err := strconv.Atoi(strIndex)
		if err != nil {
			return errors.New("field not found")
		}
		keySuffix := key[closeBracket+1:]
		return AlterFieldOfNestedStringMaps(knownType[intIndex], keySuffix, value)
	default:
		return fmt.Errorf("unexpected type: %T", knownType)
	}
}

func getResourceMapping(mapper meta.RESTMapper, obj *unstructured.Unstructured) (*meta.RESTMapping, error) {
	gvk := obj.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	return mapping, nil
}

func GetQualifiedResource(mapper meta.RESTMapper, obj *unstructured.Unstructured) (string, error) {
	mapping, err := getResourceMapping(mapper, obj)
	if err != nil {
		return "", err
	}

	if mapping.Resource.Group == "" {
		return mapping.Resource.Resource, nil
	}
	return fmt.Sprintf("%s.%s", mapping.Resource.Resource, mapping.Resource.Group), nil
}

func GetQualifiedResourceWithName(mapper meta.RESTMapper, obj *unstructured.Unstructured) (string, error) {
	mapping, err := getResourceMapping(mapper, obj)
	if err != nil {
		return "", err
	}
	objName := obj.GetName()
	if objName == "" {
		objName = obj.GetGenerateName()
	}

	if mapping.Resource.Group == "" {
		return fmt.Sprintf("%s/%s", mapping.Resource.Resource, objName), nil
	}
	return fmt.Sprintf("%s.%s/%s", mapping.Resource.Resource, mapping.Resource.Group, objName), nil
}
