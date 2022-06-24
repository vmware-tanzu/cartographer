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

// -----------------------------------------
// -- OWNER.STATUS.CONDITIONS --
// ConditionTypes
//   Workload            Deliverable
//     SupplyChainReady    DeliveryReady
//     ResourcesSubmitted  ResourcesSubmitted
//     Ready               Ready

// -- OWNER ConditionTypes

const (
	OwnerReady               = "Ready"
	ResourcesHealthy         = "ResourcesHealthy"
	WorkloadSupplyChainReady = "SupplyChainReady"
	DeliverableDeliveryReady = "DeliveryReady"
	OwnerResourcesSubmitted  = "ResourcesSubmitted"
)

// -- OWNER ConditionType - SupplyChainReady ConditionReasons

const (
	ReadySupplyChainReason                 = "Ready"
	WorkloadLabelsMissingSupplyChainReason = "WorkloadLabelsMissing"
	NotFoundSupplyChainReadyReason         = "SupplyChainNotFound"
	MultipleMatchesSupplyChainReadyReason  = "MultipleSupplyChainMatches"
)

// -- OWNER ConditionType - DeliveryReady ConditionReasons

const (
	ReadyDeliveryReason                    = "Ready"
	DeliverableLabelsMissingDeliveryReason = "DeliverableLabelsMissing"
	NotFoundDeliveryReadyReason            = "DeliveryNotFound"
	MultipleMatchesDeliveryReadyReason     = "MultipleDeliveryMatches"
)

// -- RESOURCE ConditionType - ResourceSubmitted ConditionReasons &&
// -- OWNER ConditionType - ResourcesSubmitted ConditionReasons

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

// -- RESOURCE (OWNER DELIVERABLE) ConditionType - ResourceSubmitted ConditionReasons &&
// -- OWNER DELIVERABLE ConditionType -ResourcesSubmitted ConditionReasons

const (
	DeploymentConditionNotMetResourcesSubmittedReason    = "ConditionNotMet"
	DeploymentFailedConditionMetResourcesSubmittedReason = "FailedConditionMet"
)

// -- OWNER ConditionType - ResourcesSubmitted ConditionReasons

const (
	ServiceAccountSecretErrorResourcesSubmittedReason    = "ServiceAccountSecretError"
	ResourceRealizerBuilderErrorResourcesSubmittedReason = "ResourceRealizerBuilderError"
)

// -----------------------------------------
// -- OWNER.STATUS.RESOURCE[x].CONDITIONS --
// ConditionTypes
//   ResourcesHealthy
//   ResourcesSubmitted
//   Ready

// -- RESOURCE ConditionTypes

const (
	ResourceReady     = "Ready"
	ResourceSubmitted = "ResourceSubmitted"
	ResourceHealthy   = "Healthy"
)

// -- RESOURCE ConditionType - ResourceSubmitted ConditionReasons (above)

// -----------------------------------------
// -- BLUEPRINT.STATUS.CONDITIONS --
// ConditionTypes
//   SupplyChain         Delivery
//     TemplatesReady      TemplatesReady
//     Ready               Ready

// -- BLUEPRINT ConditionTypes

const (
	BlueprintTemplatesReady = "TemplatesReady"
	BlueprintReady          = "Ready"
)

// -- BLUEPRINT ConditionType - TemplatesReady ConditionReasons

const (
	ReadyTemplatesReadyReason    = "Ready"
	NotFoundTemplatesReadyReason = "TemplatesNotFound"
)

// -- BLUEPRINT ConditionType - ResourcesHealthy True ConditionReasons

const (
	OutputAvailableResourcesHealthyReason = "OutputsAvailable"
	AlwaysHealthyResourcesHealthyReason   = "AlwaysHealthy"
)

// -- BLUEPRINT ConditionType - ResourcesHealthy Unknown ConditionReasons

const (
	NoResourceResourcesHealthyReason         = "NoResource"
	OutputNotAvailableResourcesHealthyReason = "OutputNotAvailable"
	NoStampedObjectHealthyReason             = "NoStampedObject"
	NoMatchesFulfilledReason                 = "NoMatchesFulfilled"
)

// -- BLUEPRINT ConditionType - ResourcesHealthy MultiMatch ConditionReasons

const (
	MultiMatchConditionHealthyReason = "MatchedCondition"
	MultiMatchFieldHealthyReason     = "MatchedField"
)

// -----------------------------------------
// -- RUNNABLE.STATUS.CONDITIONS --
// ConditionTypes
//   RunTemplateReady
//   Ready

// -- RUNNABLE ConditionTypes

const (
	RunnableReady          = "Ready"
	RunTemplateReady       = "RunTemplateReady"
	StampedObjectCondition = "StampedObjectCondition"
)

// -- RUNNABLE ConditionType - RunTemplateReady ConditionReasons

const (
	ReadyRunTemplateReason                            = "Ready"
	NotFoundRunTemplateReason                         = "RunTemplateNotFound"
	StampedObjectRejectedByAPIServerRunTemplateReason = "StampedObjectRejectedByAPIServer"
	OutputPathNotSatisfiedRunTemplateReason           = "OutputPathNotSatisfied"
	TemplateStampFailureRunTemplateReason             = "TemplateStampFailure"
	FailedToListCreatedObjectsReason                  = "FailedToListCreatedObjects"
	UnknownErrorReason                                = "UnknownError"
	ClientBuilderErrorResourcesSubmittedReason        = "ClientBuilderError"
	SucceededStampedObjectConditionReason             = "SucceededCondition"
	UnknownStampedObjectConditionReason               = "Unknown"
)
