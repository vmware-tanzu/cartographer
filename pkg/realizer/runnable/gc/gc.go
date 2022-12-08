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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/logger"
	"github.com/vmware-tanzu/cartographer/pkg/repository"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
)

type ByCreationTimestamp []*stamp.ExaminedObject

func (a ByCreationTimestamp) Len() int { return len(a) }
func (a ByCreationTimestamp) Less(i, j int) bool {
	return a[i].StampedObject.GetCreationTimestamp().Unix() > a[j].StampedObject.GetCreationTimestamp().Unix()
}
func (a ByCreationTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func CleanupRunnableStampedObjects(ctx context.Context, examinedObjects []*stamp.ExaminedObject, retentionPolicy v1alpha1.RetentionPolicy, repo repository.Repository) {
	log := logr.FromContextOrDiscard(ctx).WithName("runnable-stamped-object-cleanup")
	ctx = logr.NewContext(ctx, log)

	sort.Sort(ByCreationTimestamp(examinedObjects))

	var successfulFound int64
	var failedFound int64
	for _, examinedObject := range examinedObjects {
		runnableStampedObject := examinedObject.StampedObject
		runnableHealth := examinedObject.Health
		shouldDelete := false
		if runnableHealth == metav1.ConditionTrue {
			successfulFound++
			shouldDelete = successfulFound > retentionPolicy.MaxSuccessfulRuns
		} else if runnableHealth == metav1.ConditionFalse {
			failedFound++
			shouldDelete = failedFound > retentionPolicy.MaxFailedRuns
		} else {
			log.V(logger.INFO).Info("not considered for cleanup because object health has not resolved",
				"stampedObject", runnableStampedObject)
		}

		labels := runnableStampedObject.GetLabels()
		if labels["carto.run/template-lifecycle"] == "mutable" {
			shouldDelete = true
		}

		if shouldDelete {
			log.V(logger.INFO).Info("deleting runnable stamped object", "stampedObject", runnableStampedObject)
			err := repo.Delete(ctx, runnableStampedObject)
			if err != nil {
				log.Error(err, "failed to delete runnable stamped object", "stampedObject", runnableStampedObject)
			}
		}
	}
}
