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

package v1alpha1

// -- Owner
// ConditionTypes
// Workload            Deliverable
//   SupplyChainReady    DeliveryReady
//   ResourcesSubmitted  ResourcesSubmitted
//   Ready               Ready

// ConditionTypes

const (
	OwnerReady               = "Ready"
	WorkloadSupplyChainReady = "SupplyChainReady"
	DeliverableDeliveryReady = "DeliveryReady"
	OwnerResourcesSubmitted  = "ResourcesSubmitted"
)

// ConditionReasons - SupplyChainReady

const (
	ReadySupplyChainReason                 = "Ready"
	WorkloadLabelsMissingSupplyChainReason = "WorkloadLabelsMissing"
	NotFoundSupplyChainReadyReason         = "SupplyChainNotFound"
	MultipleMatchesSupplyChainReadyReason  = "MultipleSupplyChainMatches"
)

// ConditionReasons - DeliveryReady

const (
	ReadyDeliveryReason                    = "Ready"
	DeliverableLabelsMissingDeliveryReason = "DeliverableLabelsMissing"
	NotFoundDeliveryReadyReason            = "DeliveryNotFound"
	MultipleMatchesDeliveryReadyReason     = "MultipleDeliveryMatches"
)

// ConditionReasons - ResourceSubmitted &&
// ConditionReasons - ResourcesSubmitted

const (
	CompleteResourcesSubmittedReason                       = "ResourceSubmissionComplete"
	TemplateObjectRetrievalFailureResourcesSubmittedReason = "TemplateObjectRetrievalFailure"
	MissingValueAtPathResourcesSubmittedReason             = "MissingValueAtPath"
	TemplateStampFailureResourcesSubmittedReason           = "TemplateStampFailure"
	TemplateRejectedByAPIServerResourcesSubmittedReason    = "TemplateRejectedByAPIServer"
	UnknownErrorResourcesSubmittedReason                   = "UnknownError"
	ResolveTemplateOptionsErrorResourcesSubmittedReason    = "ResolveTemplateOptionsError"
	TemplateOptionsMatchErrorResourcesSubmittedReason      = "TemplateOptionsMatchError"
)

// ConditionReasons - ResourcesSubmitted

const (
	ServiceAccountSecretErrorResourcesSubmittedReason    = "ServiceAccountSecretError"
	ResourceRealizerBuilderErrorResourcesSubmittedReason = "ResourceRealizerBuilderError"
)

// ConditionReasons - ResourcesSubmitted - Deliverable

const (
	DeploymentConditionNotMetResourcesSubmittedReason    = "ConditionNotMet"
	DeploymentFailedConditionMetResourcesSubmittedReason = "FailedConditionMet"
)

// -- Owner.Status.Resource
// ConditionTypes
//   (ResourcesHealthy)
//   ResourcesSubmitted
//   Ready

// ConditionTypes

const (
	ResourceReady     = "Ready"
	ResourceSubmitted = "ResourceSubmitted"
)

// ConditionReasons - ResourceSubmitted

// -- Blueprint
// ConditionTypes
// SupplyChain         Delivery
//   TemplatesReady      TemplatesReady
//   Ready               Ready

// ConditionTypes

const (
	BlueprintTemplatesReady = "TemplatesReady"
	BlueprintReady          = "Ready"
)

// ConditionReasons - TemplatesReady

const (
	ReadyTemplatesReadyReason    = "Ready"
	NotFoundTemplatesReadyReason = "TemplatesNotFound"
)

// -- Runnable
// ConditionTypes
//   RunTempplateReady
//   Ready

// ConditionTypes

const (
	RunnableReady    = "Ready"
	RunTemplateReady = "RunTemplateReady"
)

// ConditionReasons - RunTemplateReady

const (
	ReadyRunTemplateReason                            = "Ready"
	NotFoundRunTemplateReason                         = "RunTemplateNotFound"
	StampedObjectRejectedByAPIServerRunTemplateReason = "StampedObjectRejectedByAPIServer"
	OutputPathNotSatisfiedRunTemplateReason           = "OutputPathNotSatisfied"
	TemplateStampFailureRunTemplateReason             = "TemplateStampFailure"
	FailedToListCreatedObjectsReason                  = "FailedToListCreatedObjects"
	UnknownErrorReason                                = "UnknownError"
	ClientBuilderErrorResourcesSubmittedReason        = "ClientBuilderError"
)
