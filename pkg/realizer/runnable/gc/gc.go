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

package gc

import (
	"context"
	"sort"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/eval"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
)

type ByCreationTimestamp []*unstructured.Unstructured

func (a ByCreationTimestamp) Len() int { return len(a) }
func (a ByCreationTimestamp) Less(i, j int) bool {
	return a[i].GetCreationTimestamp().Unix() > a[j].GetCreationTimestamp().Unix()
}
func (a ByCreationTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func CleanupRunnableStampedObjects(ctx context.Context, allRunnableStampedObjects []*unstructured.Unstructured, retentionPolicy v1alpha1.RetentionPolicy, repo repository.Repository) error {
	log := logr.FromContextOrDiscard(ctx).WithName("runnable-stamped-object-cleanup").WithValues("huh", "what")

	succeededConditionStatusPath := `status.conditions[?(@.type=="Succeeded")].status`
	evaluator := eval.EvaluatorBuilder()

	sort.Sort(ByCreationTimestamp(allRunnableStampedObjects))

	var successfulFound int64
	var failedFound int64
	for _, runnableStampedObject := range allRunnableStampedObjects {
		status, err := evaluator.EvaluateJsonPath(succeededConditionStatusPath, runnableStampedObject.UnstructuredContent())
		if err != nil {
			log.Error(err, "failed evaluating jsonpath to determine runnable stamped object success", "stampedObject", runnableStampedObject)
		}
		if status == "True" {
			successfulFound++
			if successfulFound > retentionPolicy.NumSuccessfulRuns {
				err = repo.Delete(context.TODO(), runnableStampedObject)
				if err != nil {
					log.Error(err, "failed to delete runnable stamped object", "stampedObject", runnableStampedObject)
				}
			}
		} else if status == "False" {
			failedFound++
			if failedFound > retentionPolicy.NumFailedRuns {
				err = repo.Delete(context.TODO(), runnableStampedObject)
				if err != nil {
					log.Error(err, "failed to delete runnable stamped object", "stampedObject", runnableStampedObject)
				}
			}
		}
	}

	return nil
}
