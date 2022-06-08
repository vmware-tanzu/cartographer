//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by diegen. DO NOT EDIT.

package dies

import (
	v1 "dies.dev/apis/meta/v1"
	json "encoding/json"
	fmtx "fmt"
	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/tests/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

var ClusterRunTemplateBlank = (&ClusterRunTemplateDie{}).DieFeed(v1alpha1.ClusterRunTemplate{})

type ClusterRunTemplateDie struct {
	v1.FrozenObjectMeta
	mutable bool
	r       v1alpha1.ClusterRunTemplate
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *ClusterRunTemplateDie) DieImmutable(immutable bool) *ClusterRunTemplateDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *ClusterRunTemplateDie) DieFeed(r v1alpha1.ClusterRunTemplate) *ClusterRunTemplateDie {
	if d.mutable {
		d.FrozenObjectMeta = v1.FreezeObjectMeta(r.ObjectMeta)
		d.r = r
		return d
	}
	return &ClusterRunTemplateDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *ClusterRunTemplateDie) DieFeedPtr(r *v1alpha1.ClusterRunTemplate) *ClusterRunTemplateDie {
	if r == nil {
		r = &v1alpha1.ClusterRunTemplate{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *ClusterRunTemplateDie) DieRelease() v1alpha1.ClusterRunTemplate {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *ClusterRunTemplateDie) DieReleasePtr() *v1alpha1.ClusterRunTemplate {
	r := d.DieRelease()
	return &r
}

// DieReleaseUnstructured returns the resource managed by the die as an unstructured object.
func (d *ClusterRunTemplateDie) DieReleaseUnstructured() runtime.Unstructured {
	r := d.DieReleasePtr()
	u, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	return &unstructured.Unstructured{
		Object: u,
	}
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *ClusterRunTemplateDie) DieStamp(fn func(r *v1alpha1.ClusterRunTemplate)) *ClusterRunTemplateDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *ClusterRunTemplateDie) DeepCopy() *ClusterRunTemplateDie {
	r := *d.r.DeepCopy()
	return &ClusterRunTemplateDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

var _ runtime.Object = (*ClusterRunTemplateDie)(nil)

func (d *ClusterRunTemplateDie) DeepCopyObject() runtime.Object {
	return d.r.DeepCopy()
}

func (d *ClusterRunTemplateDie) GetObjectKind() schema.ObjectKind {
	r := d.DieRelease()
	return r.GetObjectKind()
}

func (d *ClusterRunTemplateDie) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.r)
}

func (d *ClusterRunTemplateDie) UnmarshalJSON(b []byte) error {
	if d == ClusterRunTemplateBlank {
		return fmtx.Errorf("cannot unmarshal into the blank die, create a copy first")
	}
	if !d.mutable {
		return fmtx.Errorf("cannot unmarshal into immutable dies, create a mutable version first")
	}
	r := &v1alpha1.ClusterRunTemplate{}
	err := json.Unmarshal(b, r)
	*d = *d.DieFeed(*r)
	return err
}

// MetadataDie stamps the resource's ObjectMeta field with a mutable die.
func (d *ClusterRunTemplateDie) MetadataDie(fn func(d *v1.ObjectMetaDie)) *ClusterRunTemplateDie {
	return d.DieStamp(func(r *v1alpha1.ClusterRunTemplate) {
		d := v1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

// SpecDie stamps the resource's spec field with a mutable die.
func (d *ClusterRunTemplateDie) SpecDie(fn func(d *RunTemplateSpecDie)) *ClusterRunTemplateDie {
	return d.DieStamp(func(r *v1alpha1.ClusterRunTemplate) {
		d := RunTemplateSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}

// Spec describes the run template. More info: https://cartographer.sh/docs/latest/reference/runnable/#clusterruntemplate
func (d *ClusterRunTemplateDie) Spec(v v1alpha1.RunTemplateSpec) *ClusterRunTemplateDie {
	return d.DieStamp(func(r *v1alpha1.ClusterRunTemplate) {
		r.Spec = v
	})
}

var RunTemplateSpecBlank = (&RunTemplateSpecDie{}).DieFeed(v1alpha1.RunTemplateSpec{})

type RunTemplateSpecDie struct {
	mutable bool
	r       v1alpha1.RunTemplateSpec
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *RunTemplateSpecDie) DieImmutable(immutable bool) *RunTemplateSpecDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *RunTemplateSpecDie) DieFeed(r v1alpha1.RunTemplateSpec) *RunTemplateSpecDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &RunTemplateSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *RunTemplateSpecDie) DieFeedPtr(r *v1alpha1.RunTemplateSpec) *RunTemplateSpecDie {
	if r == nil {
		r = &v1alpha1.RunTemplateSpec{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *RunTemplateSpecDie) DieRelease() v1alpha1.RunTemplateSpec {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *RunTemplateSpecDie) DieReleasePtr() *v1alpha1.RunTemplateSpec {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *RunTemplateSpecDie) DieStamp(fn func(r *v1alpha1.RunTemplateSpec)) *RunTemplateSpecDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *RunTemplateSpecDie) DeepCopy() *RunTemplateSpecDie {
	r := *d.r.DeepCopy()
	return &RunTemplateSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// Template defines a resource template for a Kubernetes Resource or Custom Resource which is applied to the server each time the blueprint is applied. Templates support simple value interpolation using the $()$ marker format. For more information, see: https://cartographer.sh/docs/latest/templating/ You should not define the namespace for the resource - it will automatically be created in the owner namespace. If the namespace is specified and is not the owner namespace, the resource will fail to be created.
func (d *RunTemplateSpecDie) Template(v runtime.RawExtension) *RunTemplateSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunTemplateSpec) {
		r.Template = v
	})
}

// Outputs are a named list of jsonPaths that are used to gather results from the last successful object stamped by the template. E.g: 	my-output: .status.results[?(@.name=="IMAGE-DIGEST")].value Note: outputs are only filled on the runnable when the templated object has a Succeeded condition with a Status of True E.g:     status.conditions[?(@.type=="Succeeded")].status == True a runnable creating an object without a Succeeded condition (like a Job or ConfigMap) will never display an output
func (d *RunTemplateSpecDie) Outputs(v map[string]string) *RunTemplateSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunTemplateSpec) {
		r.Outputs = v
	})
}

var RunnableBlank = (&RunnableDie{}).DieFeed(v1alpha1.Runnable{})

type RunnableDie struct {
	v1.FrozenObjectMeta
	mutable bool
	r       v1alpha1.Runnable
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *RunnableDie) DieImmutable(immutable bool) *RunnableDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *RunnableDie) DieFeed(r v1alpha1.Runnable) *RunnableDie {
	if d.mutable {
		d.FrozenObjectMeta = v1.FreezeObjectMeta(r.ObjectMeta)
		d.r = r
		return d
	}
	return &RunnableDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *RunnableDie) DieFeedPtr(r *v1alpha1.Runnable) *RunnableDie {
	if r == nil {
		r = &v1alpha1.Runnable{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *RunnableDie) DieRelease() v1alpha1.Runnable {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *RunnableDie) DieReleasePtr() *v1alpha1.Runnable {
	r := d.DieRelease()
	return &r
}

// DieReleaseUnstructured returns the resource managed by the die as an unstructured object.
func (d *RunnableDie) DieReleaseUnstructured() runtime.Unstructured {
	r := d.DieReleasePtr()
	u, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	return &unstructured.Unstructured{
		Object: u,
	}
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *RunnableDie) DieStamp(fn func(r *v1alpha1.Runnable)) *RunnableDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *RunnableDie) DeepCopy() *RunnableDie {
	r := *d.r.DeepCopy()
	return &RunnableDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

var _ runtime.Object = (*RunnableDie)(nil)

func (d *RunnableDie) DeepCopyObject() runtime.Object {
	return d.r.DeepCopy()
}

func (d *RunnableDie) GetObjectKind() schema.ObjectKind {
	r := d.DieRelease()
	return r.GetObjectKind()
}

func (d *RunnableDie) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.r)
}

func (d *RunnableDie) UnmarshalJSON(b []byte) error {
	if d == RunnableBlank {
		return fmtx.Errorf("cannot unmarshal into the blank die, create a copy first")
	}
	if !d.mutable {
		return fmtx.Errorf("cannot unmarshal into immutable dies, create a mutable version first")
	}
	r := &v1alpha1.Runnable{}
	err := json.Unmarshal(b, r)
	*d = *d.DieFeed(*r)
	return err
}

// MetadataDie stamps the resource's ObjectMeta field with a mutable die.
func (d *RunnableDie) MetadataDie(fn func(d *v1.ObjectMetaDie)) *RunnableDie {
	return d.DieStamp(func(r *v1alpha1.Runnable) {
		d := v1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

// SpecDie stamps the resource's spec field with a mutable die.
func (d *RunnableDie) SpecDie(fn func(d *RunnableSpecDie)) *RunnableDie {
	return d.DieStamp(func(r *v1alpha1.Runnable) {
		d := RunnableSpecBlank.DieImmutable(false).DieFeed(r.Spec)
		fn(d)
		r.Spec = d.DieRelease()
	})
}

// StatusDie stamps the resource's status field with a mutable die.
func (d *RunnableDie) StatusDie(fn func(d *RunnableStatusDie)) *RunnableDie {
	return d.DieStamp(func(r *v1alpha1.Runnable) {
		d := RunnableStatusBlank.DieImmutable(false).DieFeed(r.Status)
		fn(d)
		r.Status = d.DieRelease()
	})
}

// Spec describes the runnable. More info: https://cartographer.sh/docs/latest/reference/runnable/#runnable
func (d *RunnableDie) Spec(v v1alpha1.RunnableSpec) *RunnableDie {
	return d.DieStamp(func(r *v1alpha1.Runnable) {
		r.Spec = v
	})
}

// Status conforms to the Kubernetes conventions: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
func (d *RunnableDie) Status(v v1alpha1.RunnableStatus) *RunnableDie {
	return d.DieStamp(func(r *v1alpha1.Runnable) {
		r.Status = v
	})
}

var RunnableSpecBlank = (&RunnableSpecDie{}).DieFeed(v1alpha1.RunnableSpec{})

type RunnableSpecDie struct {
	mutable bool
	r       v1alpha1.RunnableSpec
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *RunnableSpecDie) DieImmutable(immutable bool) *RunnableSpecDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *RunnableSpecDie) DieFeed(r v1alpha1.RunnableSpec) *RunnableSpecDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &RunnableSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *RunnableSpecDie) DieFeedPtr(r *v1alpha1.RunnableSpec) *RunnableSpecDie {
	if r == nil {
		r = &v1alpha1.RunnableSpec{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *RunnableSpecDie) DieRelease() v1alpha1.RunnableSpec {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *RunnableSpecDie) DieReleasePtr() *v1alpha1.RunnableSpec {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *RunnableSpecDie) DieStamp(fn func(r *v1alpha1.RunnableSpec)) *RunnableSpecDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *RunnableSpecDie) DeepCopy() *RunnableSpecDie {
	r := *d.r.DeepCopy()
	return &RunnableSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// RunTemplateRef identifies the run template used to produce resources for this runnable.
func (d *RunnableSpecDie) RunTemplateRef(v v1alpha1.TemplateReference) *RunnableSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunnableSpec) {
		r.RunTemplateRef = v
	})
}

// Selector refers to an additional object that the template can refer to using: $(selected)$.
func (d *RunnableSpecDie) Selector(v *v1alpha1.ResourceSelector) *RunnableSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunnableSpec) {
		r.Selector = v
	})
}

// ServiceAccountName refers to the Service account with permissions to create resources submitted by the ClusterRunTemplate.
//
// If not set, Cartographer will use the default service account in the runnable's namespace.
func (d *RunnableSpecDie) ServiceAccountName(v string) *RunnableSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunnableSpec) {
		r.ServiceAccountName = v
	})
}

// RetentionPolicy specifies how many successful and failed runs should be retained. Runs older than this (ordered by creation time) will be deleted. Setting higher values will increase memory footprint.
func (d *RunnableSpecDie) RetentionPolicy(v v1alpha1.RetentionPolicy) *RunnableSpecDie {
	return d.DieStamp(func(r *v1alpha1.RunnableSpec) {
		r.RetentionPolicy = v
	})
}

var RunnableStatusBlank = (&RunnableStatusDie{}).DieFeed(v1alpha1.RunnableStatus{})

type RunnableStatusDie struct {
	mutable bool
	r       v1alpha1.RunnableStatus
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *RunnableStatusDie) DieImmutable(immutable bool) *RunnableStatusDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *RunnableStatusDie) DieFeed(r v1alpha1.RunnableStatus) *RunnableStatusDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &RunnableStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *RunnableStatusDie) DieFeedPtr(r *v1alpha1.RunnableStatus) *RunnableStatusDie {
	if r == nil {
		r = &v1alpha1.RunnableStatus{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *RunnableStatusDie) DieRelease() v1alpha1.RunnableStatus {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *RunnableStatusDie) DieReleasePtr() *v1alpha1.RunnableStatus {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *RunnableStatusDie) DieStamp(fn func(r *v1alpha1.RunnableStatus)) *RunnableStatusDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *RunnableStatusDie) DeepCopy() *RunnableStatusDie {
	r := *d.r.DeepCopy()
	return &RunnableStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *RunnableStatusDie) ObservedGeneration(v int64) *RunnableStatusDie {
	return d.DieStamp(func(r *v1alpha1.RunnableStatus) {
		r.ObservedGeneration = v
	})
}

func (d *RunnableStatusDie) Conditions(v ...metav1.Condition) *RunnableStatusDie {
	return d.DieStamp(func(r *v1alpha1.RunnableStatus) {
		r.Conditions = v
	})
}

var TestObjBlank = (&TestObjDie{}).DieFeed(resources.TestObj{})

type TestObjDie struct {
	v1.FrozenObjectMeta
	mutable bool
	r       resources.TestObj
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *TestObjDie) DieImmutable(immutable bool) *TestObjDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *TestObjDie) DieFeed(r resources.TestObj) *TestObjDie {
	if d.mutable {
		d.FrozenObjectMeta = v1.FreezeObjectMeta(r.ObjectMeta)
		d.r = r
		return d
	}
	return &TestObjDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *TestObjDie) DieFeedPtr(r *resources.TestObj) *TestObjDie {
	if r == nil {
		r = &resources.TestObj{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *TestObjDie) DieRelease() resources.TestObj {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *TestObjDie) DieReleasePtr() *resources.TestObj {
	r := d.DieRelease()
	return &r
}

// DieReleaseUnstructured returns the resource managed by the die as an unstructured object.
func (d *TestObjDie) DieReleaseUnstructured() runtime.Unstructured {
	r := d.DieReleasePtr()
	u, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(r)
	return &unstructured.Unstructured{
		Object: u,
	}
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *TestObjDie) DieStamp(fn func(r *resources.TestObj)) *TestObjDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *TestObjDie) DeepCopy() *TestObjDie {
	r := *d.r.DeepCopy()
	return &TestObjDie{
		FrozenObjectMeta: v1.FreezeObjectMeta(r.ObjectMeta),
		mutable:          d.mutable,
		r:                r,
	}
}

var _ runtime.Object = (*TestObjDie)(nil)

func (d *TestObjDie) DeepCopyObject() runtime.Object {
	return d.r.DeepCopy()
}

func (d *TestObjDie) GetObjectKind() schema.ObjectKind {
	r := d.DieRelease()
	return r.GetObjectKind()
}

func (d *TestObjDie) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.r)
}

func (d *TestObjDie) UnmarshalJSON(b []byte) error {
	if d == TestObjBlank {
		return fmtx.Errorf("cannot unmarshal into the blank die, create a copy first")
	}
	if !d.mutable {
		return fmtx.Errorf("cannot unmarshal into immutable dies, create a mutable version first")
	}
	r := &resources.TestObj{}
	err := json.Unmarshal(b, r)
	*d = *d.DieFeed(*r)
	return err
}

// MetadataDie stamps the resource's ObjectMeta field with a mutable die.
func (d *TestObjDie) MetadataDie(fn func(d *v1.ObjectMetaDie)) *TestObjDie {
	return d.DieStamp(func(r *resources.TestObj) {
		d := v1.ObjectMetaBlank.DieImmutable(false).DieFeed(r.ObjectMeta)
		fn(d)
		r.ObjectMeta = d.DieRelease()
	})
}

func (d *TestObjDie) Spec(v resources.TestSpec) *TestObjDie {
	return d.DieStamp(func(r *resources.TestObj) {
		r.Spec = v
	})
}

func (d *TestObjDie) Status(v resources.TestStatus) *TestObjDie {
	return d.DieStamp(func(r *resources.TestObj) {
		r.Status = v
	})
}

var TestStatusBlank = (&TestStatusDie{}).DieFeed(resources.TestStatus{})

type TestStatusDie struct {
	mutable bool
	r       resources.TestStatus
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *TestStatusDie) DieImmutable(immutable bool) *TestStatusDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *TestStatusDie) DieFeed(r resources.TestStatus) *TestStatusDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &TestStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *TestStatusDie) DieFeedPtr(r *resources.TestStatus) *TestStatusDie {
	if r == nil {
		r = &resources.TestStatus{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *TestStatusDie) DieRelease() resources.TestStatus {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *TestStatusDie) DieReleasePtr() *resources.TestStatus {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *TestStatusDie) DieStamp(fn func(r *resources.TestStatus)) *TestStatusDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *TestStatusDie) DeepCopy() *TestStatusDie {
	r := *d.r.DeepCopy()
	return &TestStatusDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *TestStatusDie) ObservedGeneration(v int64) *TestStatusDie {
	return d.DieStamp(func(r *resources.TestStatus) {
		r.ObservedGeneration = v
	})
}

func (d *TestStatusDie) Conditions(v ...metav1.Condition) *TestStatusDie {
	return d.DieStamp(func(r *resources.TestStatus) {
		r.Conditions = v
	})
}

var TestSpecBlank = (&TestSpecDie{}).DieFeed(resources.TestSpec{})

type TestSpecDie struct {
	mutable bool
	r       resources.TestSpec
}

// DieImmutable returns a new die for the current die's state that is either mutable (`false`) or immutable (`true`).
func (d *TestSpecDie) DieImmutable(immutable bool) *TestSpecDie {
	if d.mutable == !immutable {
		return d
	}
	d = d.DeepCopy()
	d.mutable = !immutable
	return d
}

// DieFeed returns a new die with the provided resource.
func (d *TestSpecDie) DieFeed(r resources.TestSpec) *TestSpecDie {
	if d.mutable {
		d.r = r
		return d
	}
	return &TestSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

// DieFeedPtr returns a new die with the provided resource pointer. If the resource is nil, the empty value is used instead.
func (d *TestSpecDie) DieFeedPtr(r *resources.TestSpec) *TestSpecDie {
	if r == nil {
		r = &resources.TestSpec{}
	}
	return d.DieFeed(*r)
}

// DieRelease returns the resource managed by the die.
func (d *TestSpecDie) DieRelease() resources.TestSpec {
	if d.mutable {
		return d.r
	}
	return *d.r.DeepCopy()
}

// DieReleasePtr returns a pointer to the resource managed by the die.
func (d *TestSpecDie) DieReleasePtr() *resources.TestSpec {
	r := d.DieRelease()
	return &r
}

// DieStamp returns a new die with the resource passed to the callback function. The resource is mutable.
func (d *TestSpecDie) DieStamp(fn func(r *resources.TestSpec)) *TestSpecDie {
	r := d.DieRelease()
	fn(&r)
	return d.DieFeed(r)
}

// DeepCopy returns a new die with equivalent state. Useful for snapshotting a mutable die.
func (d *TestSpecDie) DeepCopy() *TestSpecDie {
	r := *d.r.DeepCopy()
	return &TestSpecDie{
		mutable: d.mutable,
		r:       r,
	}
}

func (d *TestSpecDie) Foo(v string) *TestSpecDie {
	return d.DieStamp(func(r *resources.TestSpec) {
		r.Foo = v
	})
}

func (d *TestSpecDie) Value(v runtime.RawExtension) *TestSpecDie {
	return d.DieStamp(func(r *resources.TestSpec) {
		r.Value = v
	})
}
