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

package gc_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/cartographer/pkg/apis/v1alpha1"
	"github.com/vmware-tanzu/cartographer/pkg/realizer/runnable/gc"
	"github.com/vmware-tanzu/cartographer/pkg/repository/repositoryfakes"
	"github.com/vmware-tanzu/cartographer/pkg/stamp"
	"github.com/vmware-tanzu/cartographer/pkg/utils"
)

func MakeRunnableStampedObject(status, name, creationTimeStamp string) *unstructured.Unstructured {
	yamlString := utils.HereYaml(`
					---
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  creationTimestamp: ` + creationTimeStamp + `
					  name: ` + name + `
					status:
					  conditions:
						- type: Succeeded
						  status: "` + status + `"
	`)
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(yamlString), obj)
	Expect(err).NotTo(HaveOccurred())
	return obj
}

var _ = Describe("CleanupRunnableStampedObjects", func() {
	var (
		repo               *repositoryfakes.FakeRepository
		allExaminedObjects []*stamp.ExaminedObject
		retentionPolicy    v1alpha1.RetentionPolicy
		ctx                context.Context
		out                *Buffer
	)

	BeforeEach(func() {
		out = NewBuffer()
		logger := zap.New(zap.WriteTo(out))
		ctx = logr.NewContext(context.Background(), logger)

		allExaminedObjects = []*stamp.ExaminedObject{
			{
				StampedObject: MakeRunnableStampedObject("True", "RecentSuccessRetainedByPolicy1", "2022-01-11T17:00:07Z"),
				Health:        metav1.ConditionTrue,
			},
			{
				StampedObject: MakeRunnableStampedObject("True", "MostRecentSuccess", "2022-01-12T17:00:07Z"),
				Health:        metav1.ConditionTrue,
			},
			{
				StampedObject: MakeRunnableStampedObject("False", "MostRecentFailure", "2022-01-12T17:00:07Z"),
				Health:        metav1.ConditionFalse,
			},
			{
				StampedObject: MakeRunnableStampedObject("False", "RecentFailureRetainedByPolicy2", "2022-01-11T17:00:07Z"),
				Health:        metav1.ConditionFalse,
			},
			{
				StampedObject: MakeRunnableStampedObject("True", "RecentSuccessRetainedByPolicy2", "2022-01-10T17:00:07Z"),
				Health:        metav1.ConditionTrue,
			},
		}

		repo = &repositoryfakes.FakeRepository{}

		retentionPolicy = v1alpha1.RetentionPolicy{MaxFailedRuns: 2, MaxSuccessfulRuns: 3}
	})

	It("should not error, but log a warning, when a stamped object that doesnt have a Succeeded status is handled", func() {
		yamlString := utils.HereYaml(`
					---
					apiVersion: test.run/v1alpha1
					kind: TestObj
					metadata:
					  name: ThingWithoutSucceededCondition
					  creationTimestamp: 2022-01-30T17:00:07Z
		`)
		objectWithoutSucceededStatus := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(yamlString), objectWithoutSucceededStatus)
		Expect(err).NotTo(HaveOccurred())

		examinedObjToStillExist := stamp.ExaminedObject{
			StampedObject: objectWithoutSucceededStatus,
			Health:        metav1.ConditionUnknown,
		}

		successfulRunnableStampedObjectToBeDeleted := MakeRunnableStampedObject("True", "RecentSuccessToBeDeleted1", "2022-01-09T17:00:07Z")

		examinedObjToBeDeleted := stamp.ExaminedObject{
			StampedObject: successfulRunnableStampedObjectToBeDeleted,
			Health:        metav1.ConditionTrue,
		}

		allExaminedObjects = append([]*stamp.ExaminedObject{&examinedObjToStillExist, &examinedObjToBeDeleted}, allExaminedObjects...)

		gc.CleanupRunnableStampedObjects(ctx, allExaminedObjects, retentionPolicy, repo)

		Expect(repo.DeleteCallCount()).To(Equal(1))
		_, deletedObject1 := repo.DeleteArgsForCall(0)
		Expect(deletedObject1).To(Equal(successfulRunnableStampedObjectToBeDeleted))
	})

	Context("when runnable stamped objects outside the retention policy are processed", func() {
		var (
			successfulRunnableStampedObjectToBeDeleted1 *unstructured.Unstructured
			successfulRunnableStampedObjectToBeDeleted2 *unstructured.Unstructured
			failedRunnableStampedObjectToBeDeleted1     *unstructured.Unstructured
			failedRunnableStampedObjectToBeDeleted2     *unstructured.Unstructured
		)

		BeforeEach(func() {
			successfulRunnableStampedObjectToBeDeleted1 = MakeRunnableStampedObject("True", "RecentSuccessToBeDeleted1", "2022-01-09T17:00:07Z")
			successfulRunnableStampedObjectToBeDeleted2 = MakeRunnableStampedObject("True", "RecentSuccessToBeDeleted2", "2022-01-08T17:00:07Z")
			failedRunnableStampedObjectToBeDeleted1 = MakeRunnableStampedObject("False", "RecentFailureToBeDeleted1", "2022-01-10T17:00:07Z")
			failedRunnableStampedObjectToBeDeleted2 = MakeRunnableStampedObject("False", "RecentFailureToBeDeleted2", "2022-01-09T17:00:07Z")

			//ensure dates are out of order for the items to be deleted
			allExaminedObjects = append(allExaminedObjects,
				&stamp.ExaminedObject{
					StampedObject: successfulRunnableStampedObjectToBeDeleted1,
					Health:        metav1.ConditionTrue,
				},
				&stamp.ExaminedObject{
					StampedObject: failedRunnableStampedObjectToBeDeleted1,
					Health:        metav1.ConditionFalse,
				},
			)
			allExaminedObjects = append([]*stamp.ExaminedObject{
				{
					StampedObject: successfulRunnableStampedObjectToBeDeleted2,
					Health:        metav1.ConditionTrue,
				},
				{
					StampedObject: failedRunnableStampedObjectToBeDeleted2,
					Health:        metav1.ConditionFalse,
				},
			},
				allExaminedObjects...,
			)
		})

		It("continues processing all elements and logs an error if deleting a runnable stamped object fails", func() {
			repo.DeleteReturns(errors.New("deleting is hard"))
			gc.CleanupRunnableStampedObjects(ctx, allExaminedObjects, retentionPolicy, repo)

			Expect(repo.DeleteCallCount()).To(Equal(4))
			Expect(out).To(Say("failed to delete runnable stamped object.*RecentFailureToBeDeleted1.*deleting is hard"))
			Expect(out).To(Say("failed to delete runnable stamped object.*RecentFailureToBeDeleted2.*deleting is hard"))
			Expect(out).To(Say("failed to delete runnable stamped object.*RecentSuccessToBeDeleted1.*deleting is hard"))
			Expect(out).To(Say("failed to delete runnable stamped object.*RecentSuccessToBeDeleted2.*deleting is hard"))
		})

		It("deletes successful and failed runnable stamped objects according to retention policy", func() {
			gc.CleanupRunnableStampedObjects(ctx, allExaminedObjects, retentionPolicy, repo)

			Expect(repo.DeleteCallCount()).To(Equal(4))
			_, deletedObject1 := repo.DeleteArgsForCall(0)
			_, deletedObject2 := repo.DeleteArgsForCall(1)
			_, deletedObject3 := repo.DeleteArgsForCall(2)
			_, deletedObject4 := repo.DeleteArgsForCall(3)
			Expect([]*unstructured.Unstructured{
				deletedObject1,
				deletedObject2,
				deletedObject3,
				deletedObject4}).To(ConsistOf(
				successfulRunnableStampedObjectToBeDeleted1,
				successfulRunnableStampedObjectToBeDeleted2,
				failedRunnableStampedObjectToBeDeleted1,
				failedRunnableStampedObjectToBeDeleted2,
			))
		})

		It("ignores runnable stamped objects that have not succeeded or failed", func() {
			failedRunnableStampedObjectToBeIgnored1 := MakeRunnableStampedObject("Unknown", "RecentFailureToBeDeleted1", "2022-01-10T17:00:07Z")
			failedRunnableStampedObjectToBeIgnored2 := MakeRunnableStampedObject("Unknown", "RecentFailureToBeDeleted2", "2022-01-09T17:00:07Z")

			//ensure dates are out of order for the items to be deleted
			allExaminedObjects = append(allExaminedObjects, &stamp.ExaminedObject{
				StampedObject: failedRunnableStampedObjectToBeIgnored1,
				Health:        metav1.ConditionUnknown,
			})
			allExaminedObjects = append([]*stamp.ExaminedObject{
				{
					StampedObject: failedRunnableStampedObjectToBeIgnored2,
					Health:        metav1.ConditionUnknown,
				}}, allExaminedObjects...)

			gc.CleanupRunnableStampedObjects(ctx, allExaminedObjects, retentionPolicy, repo)

			Expect(repo.DeleteCallCount()).To(Equal(4))
			_, deletedObject1 := repo.DeleteArgsForCall(0)
			_, deletedObject2 := repo.DeleteArgsForCall(1)
			_, deletedObject3 := repo.DeleteArgsForCall(2)
			_, deletedObject4 := repo.DeleteArgsForCall(3)
			Expect([]*unstructured.Unstructured{
				deletedObject1,
				deletedObject2,
				deletedObject3,
				deletedObject4}).NotTo(ContainElements(failedRunnableStampedObjectToBeIgnored1, failedRunnableStampedObjectToBeIgnored2))
			Expect(out).To(Say("deleting runnable stamped object"))
		})
	})
})
