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

package events

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

const QualifiedResourceNameToken = "%Q"

//go:generate go run -modfile ../../hack/tools/go.mod github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate k8s.io/client-go/tools/record.EventRecorder
//counterfeiter:generate k8s.io/apimachinery/pkg/api/meta.RESTMapper
//counterfeiter:generate . OwnerEventRecorder
type OwnerEventRecorder interface {
	// Event uses an EventRecorder to record an event against this OwnerEventRecorder's owner object
	Event(eventtype, reason, message string)

	// Eventf is just like Event, but with Sprintf for the message field.
	Eventf(eventtype, reason, messageFmt string, args ...interface{})

	// AnnotatedEventf is just like eventf, but with annotations attached
	AnnotatedEventf(annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{})

	// ResourceEventf is just like Eventf, but the token %Q will be substituted with the qualified name of the provided unstructured resource
	ResourceEventf(eventtype, reason, messageFmt string, resource *unstructured.Unstructured, args ...interface{})
}

func FromEventRecorder(rec record.EventRecorder, ownerObj runtime.Object, mapper meta.RESTMapper, log logr.Logger) OwnerEventRecorder {
	return ownerEventRecorder{
		obj:    ownerObj,
		rec:    rec,
		mapper: mapper,
		log:    log,
	}
}

// contextKey is how we find OwnerEventRecorder in a context.Context.
type contextKey struct{}

type ownerEventRecorder struct {
	obj    runtime.Object
	rec    record.EventRecorder
	mapper meta.RESTMapper
	log    logr.Logger
}

func (o ownerEventRecorder) Event(eventtype, reason, message string) {
	o.rec.Event(o.obj, eventtype, reason, message)
}

func (o ownerEventRecorder) Eventf(eventtype, reason, messageFmt string, args ...interface{}) {
	o.rec.Eventf(o.obj, eventtype, reason, messageFmt, args...)
}

func (o ownerEventRecorder) ResourceEventf(eventtype, reason, messageFmt string, resource *unstructured.Unstructured, args ...interface{}) {
	qualifiedResourceName, err := utils.GetQualifiedResourceName(o.mapper, resource)
	if err != nil {
		o.log.V(logger.DEBUG).Error(err, "cannot find rest mapping for resource", "resource", resource)
		return
	}
	messageFmt = strings.ReplaceAll(messageFmt, QualifiedResourceNameToken, qualifiedResourceName)
	o.rec.Eventf(o.obj, eventtype, reason, messageFmt, args...)
}

func (o ownerEventRecorder) AnnotatedEventf(annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	o.rec.AnnotatedEventf(o.obj, annotations, eventtype, reason, messageFmt, args...)
}

// FromContextOrDie returns a OwnerEventRecorder from ctx.  If no OwnerEventRecorder is found, this
// panics
func FromContextOrDie(ctx context.Context) OwnerEventRecorder {
	if v, ok := ctx.Value(contextKey{}).(OwnerEventRecorder); ok {
		return v
	}

	panic("couldn't get owner event recorder from context")
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided OwnerEventRecorder.
func NewContext(ctx context.Context, rec OwnerEventRecorder) context.Context {
	return context.WithValue(ctx, contextKey{}, rec)
}
