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
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasttemplate"
	yamlv3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InterpolateResource(
	parent client.Object,
	data map[string]interface{},
	resource string,
) (*unstructured.Unstructured, error) {

	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode([]byte(resource), nil, obj)
	if err != nil {
		return nil, fmt.Errorf("decoding template to unstructured: %w", err)
	}

	if err := InterpolateInterface(data, obj.UnstructuredContent()); err != nil {
		return nil, fmt.Errorf("interpolate interface: %w", err)
	}

	hashCopy := obj.DeepCopy()
	hashObj := hashCopy.UnstructuredContent()

	delete(hashObj, "metadata")
	delete(hashObj, "kind")
	delete(hashObj, "apiVersion")

	objectDigest, err := digest(hashObj, parent.GetName())
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels["carto.run/digest"] = objectDigest

	obj.SetNamespace(parent.GetNamespace()) // TODO maybe .. not? not sure
	obj.SetLabels(labels)
	obj.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         parent.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:               parent.GetObjectKind().GroupVersionKind().Kind,
			Name:               parent.GetName(),
			UID:                parent.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	})

	return obj, nil
}

func InterpolateInterface(variables map[string]interface{}, iface interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(iface))
	if !rv.IsValid() {
		return nil
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if err := InterpolateSlice(rv, variables); err != nil {
			return fmt.Errorf("interpolate slice: %w", err)
		}
	case reflect.Map:
		if err := InterpolateMap(rv, variables); err != nil {
			return fmt.Errorf("interpolate map: %w", err)
		}
	}

	return nil
}

func InterpolateSlice(slice reflect.Value, variables map[string]interface{}) error {
	for idx := 0; idx < slice.Len(); idx++ {
		inner := slice.Index(idx).Elem()
		if inner.Kind() != reflect.String {
			if err := InterpolateInterface(variables, slice.Index(idx).Interface()); err != nil {
				return fmt.Errorf("interpolate interface: %w", err)
			}

			continue
		}

		value, err := InterpolateStringValue(inner.String(), variables)
		if err != nil {
			return fmt.Errorf("interpolate field: %w", err)
		}

		slice.Index(idx).Set(*value)
	}

	return nil
}

func InterpolateMap(mmap reflect.Value, variables map[string]interface{}) error {
	for _, key := range mmap.MapKeys() {
		inner := mmap.MapIndex(key).Elem()
		if inner.Kind() != reflect.String {
			if err := InterpolateInterface(variables, mmap.MapIndex(key).Interface()); err != nil {
				return fmt.Errorf("interpolate interface: %w", err)
			}

			continue
		}

		value, err := InterpolateStringValue(inner.String(), variables)
		if err != nil {
			return fmt.Errorf("interpolate string value: %w", err)
		}

		mmap.SetMapIndex(key, *value)
	}

	return nil
}

func InterpolateStringValue(field string, variables map[string]interface{}) (*reflect.Value, error) {
	interpolated, err := InterpolateString(field, variables)
	if err != nil {
		return nil, fmt.Errorf("interpolate string: %w", err)
	}

	value, err := StringValue(interpolated)
	if err != nil {
		return nil, fmt.Errorf("string value: %w", err)
	}

	return value, nil
}

func StringValue(str string) (*reflect.Value, error) {
	v := new(interface{})
	if err := yamlv3.Unmarshal([]byte(str), v); err != nil {
		return nil, fmt.Errorf("unmarshal to iface: %w", err)
	}

	value := reflect.ValueOf(*v)
	if value.Kind() == reflect.String {
		value = reflect.ValueOf(str)
	}
	if value.Kind() == reflect.Int {
		i := (*v).(int)
		int64v := int64(i)
		value = reflect.ValueOf(int64v)
	}

	return &value, nil
}

func InterpolateString(str string, data map[string]interface{}) (string, error) {
	result, err := fasttemplate.ExecuteFuncStringWithErr(str,
		`$(`, `)$`,
		func(w io.Writer, tag string) (int, error) {
			val, err := InterpolateTag(data, tag)
			if err != nil {
				return 0, fmt.Errorf("interpolate tag '%s': %w", tag, err)
			}

			return w.Write([]byte(val))
		},
	)
	if err != nil {
		return "", fmt.Errorf("execute func string with err: %w", err)
	}

	return result, nil
}

func digest(o ...interface{}) (string, error) {
	h := md5.New()

	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	_, err := printer.Fprintf(h, "%#v", o)
	if err != nil {
		return "", fmt.Errorf("printer Fprintf: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func InterpolateTag(data map[string]interface{}, tag string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("must not be empty")
	}

	parts := strings.Split(tag, ".")
	obj, found := data[parts[0]]
	if !found {
		return "", fmt.Errorf("key '%s' not found", parts[0])
	}

	objJson, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	if len(parts) == 1 {
		return string(objJson), nil
	}

	qry := strings.Join(parts[1:], ".")

	result := gjson.GetBytes(objJson, qry)
	if !result.Exists() {
		return "", fmt.Errorf("query '%s' found no results", qry)
	}

	return result.String(), nil
}

func Interpolate(data map[string]interface{}, str string) (string, error) {
	result, err := fasttemplate.ExecuteFuncStringWithErr(str,
		`$(`, `)$`,
		func(w io.Writer, tag string) (int, error) {
			val, err := InterpolateTag(data, tag)
			if err != nil {
				return 0, fmt.Errorf("interpolate tag '%s': %w", tag, err)
			}

			return w.Write([]byte(val))
		},
	)
	if err != nil {
		return "", fmt.Errorf("execute func string with err: %w", err)
	}

	return result, nil
}
