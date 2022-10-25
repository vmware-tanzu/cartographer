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

package events_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/vmware-tanzu/cartographer/pkg/events"
	"github.com/vmware-tanzu/cartographer/pkg/events/eventsfakes"
)

var _ = Describe("OwnerRecorder", func() {
	var (
		rec            events.OwnerEventRecorder
		fakeRecorder   *eventsfakes.FakeEventRecorder
		fakeMapper     *eventsfakes.FakeRESTMapper
		out            *Buffer
		ownerObject    *unstructured.Unstructured
		resourceObject *unstructured.Unstructured
	)

	BeforeEach(func() {
		fakeRecorder = &eventsfakes.FakeEventRecorder{}

		ownerObject = &unstructured.Unstructured{}
		ownerObject.SetName("the-owner")

		resourceObject = &unstructured.Unstructured{}
		resourceObject.SetName("the-resource")
		resourceObject.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "example.com",
			Version: "v1",
			Kind:    "foo",
		})

		fakeMapper = &eventsfakes.FakeRESTMapper{}

		out = NewBuffer()
		log := zap.New(zap.WriteTo(out))

		rec = events.FromEventRecorder(fakeRecorder, ownerObject, fakeMapper, log)
	})

	Describe("ResourceEventf", func() {
		It("records an event against the owner, substituting %Q in the format string with the qualified name of the unstructured resource derived from the REST mapping", func() {
			fakeMapper.RESTMappingReturns(&meta.RESTMapping{
				Resource: schema.GroupVersionResource{
					Group:    "EXAMPLE.COM",
					Version:  "v1",
					Resource: "FOO",
				},
				GroupVersionKind: resourceObject.GroupVersionKind(),
				Scope:            nil,
			}, nil)

			rec.ResourceEventf("Normal", "EggsHatched", "Resource [%Q] has hatched, looks like we'll have %d more %s!", resourceObject, 1, "hen")

			Expect(fakeMapper.RESTMappingCallCount()).To(Equal(1))
			groupKind, mappingArgs := fakeMapper.RESTMappingArgsForCall(0)
			Expect(groupKind).To(Equal(resourceObject.GroupVersionKind().GroupKind()))
			Expect(mappingArgs).To(Equal([]string{resourceObject.GroupVersionKind().Version}))

			Expect(fakeRecorder.EventfCallCount()).To(Equal(1))
			eventObject, eventType, reason, messageFormat, eventMessageArgs := fakeRecorder.EventfArgsForCall(0)
			Expect(eventObject).To(Equal(ownerObject))
			Expect(eventType).To(Equal("Normal"))
			Expect(reason).To(Equal("EggsHatched"))
			Expect(messageFormat).To(Equal("Resource [FOO.EXAMPLE.COM/the-resource] has hatched, looks like we'll have %d more %s!"))
			Expect(eventMessageArgs).To(Equal([]interface{}{1, "hen"}))
		})

		Context("REST mapper fails to resolve the object's GVK", func() {
			BeforeEach(func() {
				fakeMapper.RESTMappingReturns(nil, errors.New("mapping is hard"))
			})

			It("does not record any events and logs the error", func() {
				rec.ResourceEventf("Normal", "EggsHatched", "Resource [%Q] has hatched, looks like we'll have %d more %s!", resourceObject, 1, "hen")

				Expect(fakeRecorder.Invocations()).To(BeEmpty())
				Expect(out).To(Say(`cannot find rest mapping for resource.*"apiVersion":"example.com/v1".*"kind":"foo".*"name":"the-resource"`))
			})
		})

		It("handles nil stamped object gracefully", func() {
			rec.ResourceEventf("Normal", "EggsNil", "Resource [%Q] might panic without a nil check!", nil)

			Expect(fakeRecorder.EventfCallCount()).To(Equal(1))
			eventObject, eventType, reason, messageFormat, eventMessageArgs := fakeRecorder.EventfArgsForCall(0)
			Expect(eventObject).To(Equal(ownerObject))
			Expect(eventType).To(Equal("Normal"))
			Expect(reason).To(Equal("EggsNil"))
			Expect(messageFormat).To(Equal(`Resource [<nil>] might panic without a nil check!`))
			Expect(eventMessageArgs).To(BeEmpty())
		})
	})
})
