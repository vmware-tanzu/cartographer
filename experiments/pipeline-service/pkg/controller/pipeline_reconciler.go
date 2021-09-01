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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"
	"github.com/tidwall/gjson"
	yamlv3 "gopkg.in/yaml.v3"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	syaml "sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/experiments/pipeline-service/pkg/apis/v1alpha1"
)

type PipelineReconciler struct {
	Client client.Client
	Logger logr.Logger
}

func (r *PipelineReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {
	logger := r.Logger.WithValues("name", req.Name, "namespace", req.Namespace)

	logger.Info("started")
	defer logger.Info("finished")

	pipeline, err := r.GetPipeline(ctx, req.Name, req.Namespace)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("get pipeline: %w", err)
	}

	pipelineToReconcile := pipeline.DeepCopy()

	if err := r.ReconcilePipeline(ctx, pipelineToReconcile); err != nil {
		logger.Error(err, "reconciliation errored")

		pipelineToReconcile.Status.Conditions = []metav1.Condition{
			{
				Type:               "Errored",
				LastTransitionTime: metav1.Now(),
				Reason:             "Errored",
				Message:            err.Error(),
				Status:             "True",
			},
		}

		if r.hasStatusChanged(pipeline, pipelineToReconcile) {
			updateErr := r.Client.Status().Update(ctx, pipelineToReconcile)
			if updateErr != nil {
				logger.Error(updateErr, "update error")
			}
		}

		return ctrl.Result{}, err
	}

	if r.hasStatusChanged(pipeline, pipelineToReconcile) {
		if err := r.Client.Status().Update(ctx, pipelineToReconcile); err != nil {
			return ctrl.Result{}, fmt.Errorf("update status: %w", err)
		}
	}

	return ctrl.Result{
		RequeueAfter: 5 * time.Second,
	}, nil
}

func (r *PipelineReconciler) ReconcilePipeline(
	ctx context.Context,
	pipeline *v1alpha1.Pipeline,
) error {
	stampingContext, err := r.BuildStampingContext(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("build stamping context: %w", err)
	}

	runTemplate, err := r.GetRunTemplate(ctx, pipeline.Spec.RunTemplateName, pipeline.Namespace)
	if err != nil {
		return fmt.Errorf("get runtemplate: %w", err)
	}

	invocationObj, err := r.StampInvocationObj(ctx, pipeline, runTemplate, stampingContext)
	if err != nil {
		return fmt.Errorf("stamp invocation obj: %w", err)
	}

	err = r.FindOrCreate(ctx, invocationObj)
	if err != nil {
		return fmt.Errorf("find or create: %w", err)
	}

	err = r.FinalizePipelineStatus(ctx, pipeline, runTemplate, invocationObj)
	if err != nil {
		return fmt.Errorf("finalize pipeline status: %w", err)
	}

	return nil
}

func (r *PipelineReconciler) FinalizePipelineStatus(
	ctx context.Context,
	pipeline *v1alpha1.Pipeline,
	runTemplate *v1alpha1.RunTemplate,
	invocationObj *unstructured.Unstructured,
) error {
	transitionTime := metav1.NewTime(time.Date(2020, time.February, 1, 1, 1, 1, 1, time.Local))

	invocationStatus, err := r.ResourceCompletionStatus(ctx, invocationObj, runTemplate.Spec.Completion)
	if err != nil {
		return fmt.Errorf("resource completed: %w", err)
	}

	switch invocationStatus {
	case "Succeeded":
		latestOutputs, err := r.CollectOutputs(ctx, invocationObj, runTemplate.Spec.Outputs)
		if err != nil {
			return fmt.Errorf("collect outputs: %w", err)
		}

		b, err := json.Marshal(latestOutputs)
		if err != nil {
			return fmt.Errorf("marshal latest outputs: %w", err)
		}

		pipeline.Status.LatestOutputs = runtime.RawExtension{Raw: b}
		pipeline.Status.LatestInputs = pipeline.Spec.Inputs
		pipeline.Status.Conditions = []metav1.Condition{
			{
				LastTransitionTime: transitionTime,
				Reason:             "RunSucceeded",
				Status:             metav1.ConditionTrue,
				Type:               "Succeeded",
			},
		}
		return nil

	case "Failed":
		pipeline.Status.Conditions = []metav1.Condition{
			{
				LastTransitionTime: transitionTime,
				Reason:             "RunFailed",
				Status:             metav1.ConditionFalse,
				Type:               "Succeeded",
			},
		}
		return nil

	case "Unknown":
		pipeline.Status.Conditions = []metav1.Condition{
			{
				LastTransitionTime: transitionTime,
				Reason:             "RunRunning",
				Status:             metav1.ConditionUnknown,
				Type:               "Unknown",
			},
		}

		return nil
	}

	return fmt.Errorf("unexpected invocation status '%s'", invocationStatus)
}

func (r *PipelineReconciler) BuildStampingContext(
	ctx context.Context,
	pipeline *v1alpha1.Pipeline,
) (map[string]interface{}, error) {
	stamperContext := map[string]interface{}{}

	if pipeline.Spec.Selector.Resource.Kind != "" {
		matchingObject, err := r.FindMatchingObject(ctx, pipeline)
		if err != nil {
			return nil, fmt.Errorf("find matching obj: %w", err)
		}

		stamperContext["selected"] = matchingObject
	}

	b, err := json.Marshal(pipeline)
	if err != nil {
		return nil, fmt.Errorf("marshal pipeline: %w", err)
	}

	unstructuredPipeline := map[string]interface{}{}
	if err := json.Unmarshal(b, &unstructuredPipeline); err != nil {
		return nil, fmt.Errorf("unmarshal pipeline: %w", err)
	}

	stamperContext["pipeline"] = unstructuredPipeline

	return stamperContext, nil
}

func (r *PipelineReconciler) StampInvocationObj(
	ctx context.Context,
	pipeline *v1alpha1.Pipeline,
	runTemplate *v1alpha1.RunTemplate,
	interpolationData map[string]interface{},
) (*unstructured.Unstructured, error) {
	resource, err := syaml.JSONToYAML(runTemplate.Spec.Template.Raw)
	if err != nil {
		return nil, fmt.Errorf("json to yaml: %w", err)
	}

	obj, err := InterpolateResource(pipeline, interpolationData, string(resource))
	if err != nil {
		return nil, fmt.Errorf("interpolate resource: %w", err)
	}

	return obj, nil
}

func (r *PipelineReconciler) FindMatchingObject(
	ctx context.Context,
	pipeline *v1alpha1.Pipeline,
) (*unstructured.Unstructured, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.FromAPIVersionAndKind(
		pipeline.Spec.Selector.Resource.APIVersion,
		pipeline.Spec.Selector.Resource.Kind,
	))

	if err := r.Client.List(ctx, list,
		client.InNamespace(pipeline.Namespace),
		client.MatchingLabels(pipeline.Spec.Selector.MatchingLabels),
		client.Limit(1),
	); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	if len(list.Items) != 1 {
		return nil, fmt.Errorf("expected list w/ a single item, got %d", len(list.Items))
	}

	obj := &unstructured.Unstructured{}
	*obj = list.Items[0]

	return obj, nil
}

func (r *PipelineReconciler) CollectOutputs(
	ctx context.Context,
	obj *unstructured.Unstructured,
	outputRules []v1alpha1.RunTemplateOutput,
) (map[string]interface{}, error) {
	res := map[string]interface{}{}

	var err error
	for _, rule := range outputRules {
		res[rule.Name], err = r.CollectOutput(ctx, obj, rule)
		if err != nil {
			return nil, fmt.Errorf("collect output: %w", err)
		}
	}

	return res, nil
}

func (r *PipelineReconciler) CollectOutput(
	ctx context.Context,
	obj *unstructured.Unstructured,
	rule v1alpha1.RunTemplateOutput,
) (interface{}, error) {
	v, err := Interpolate(obj.UnstructuredContent(), rule.Path)
	if err != nil {
		return nil, fmt.Errorf("interpol: %w", err)
	}

	marshalledV := new(interface{})
	if err := yamlv3.Unmarshal([]byte(v), marshalledV); err != nil {
		return nil, fmt.Errorf("unmarshal retrieved data: %w", err)
	}

	return *marshalledV, nil
}

func (r *PipelineReconciler) ResourceCompletionStatus(
	ctx context.Context,
	obj *unstructured.Unstructured,
	completions v1alpha1.Completion,
) (string, error) {
	if completions.Failed.Key == "" && completions.Succeeded.Key == "" {
		return "Suceeded", nil
	}

	manifest, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}

	if completions.Failed.Key != "" {
		if gjson.GetBytes(manifest, completions.Failed.Key).String() == completions.Failed.Value {
			return "Failed", nil
		}
	}

	if completions.Succeeded.Key != "" {
		if gjson.GetBytes(manifest, completions.Succeeded.Key).String() == completions.Succeeded.Value {
			return "Succeeded", nil
		}
	}

	return "Unknown", nil
}

func (r *PipelineReconciler) FindOrCreate(ctx context.Context, obj *unstructured.Unstructured) error {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(obj.GroupVersionKind())

	if err := r.Client.List(ctx, list,
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels(obj.GetLabels()),
		client.Limit(1),
	); err != nil {
		return fmt.Errorf("list: %w", err)
	}

	if len(list.Items) > 0 {
		*obj = list.Items[0]
		return nil
	}

	if err := r.Client.Create(ctx, obj); err != nil {
		spew.Dump(obj)
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (r *PipelineReconciler) ListRunTemplates(
	ctx context.Context,
	opts ...client.ListOption,
) ([]*v1alpha1.RunTemplate, error) {
	list := &v1alpha1.RunTemplateList{}
	if err := r.Client.List(ctx, list, opts...); err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	res := []*v1alpha1.RunTemplate{}
	for _, item := range list.Items {
		item := item
		res = append(res, &item)
	}

	return res, nil
}

func (r *PipelineReconciler) GetRunTemplate(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.RunTemplate, error) {
	obj := &v1alpha1.RunTemplate{}

	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get '%s/%s': %w", namespace, name, err)
	}

	return obj, nil
}

func (r *PipelineReconciler) GetPipeline(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.Pipeline, error) {
	obj := &v1alpha1.Pipeline{}

	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get '%s/%s': %w", namespace, name, err)
	}

	return obj, nil
}

func (r *PipelineReconciler) hasStatusChanged(
	before *v1alpha1.Pipeline,
	after *v1alpha1.Pipeline,
) bool {
	if len(before.Status.Conditions) == 0 || len(after.Status.Conditions) == 0 {
		return true
	}

	digestBefore, err := digest(before.Status)
	if err != nil {
		panic(fmt.Errorf("digest status before: %w", err))
	}

	digestAfter, err := digest(after.Status)
	if err != nil {
		panic(fmt.Errorf("digest status before: %w", err))
	}

	return digestBefore != digestAfter
}
