# Meta
[meta]: #meta
- Name: Allow Resources to Report Status
- Start Date: 2022-03-22
- Author(s): @emmjohnson, @martyspiewak
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: (put "N/A" unless this replaces an existing RFC, then link to that RFC)

# Summary
[summary]: #summary

The RFC proposes modifying the existing templates to allow template authors to indicate how Cartographer can read the state of the resource so that it can be reflected on the owner status.

# Motivation
[motivation]: #motivation

- Why should we do this?
  - Currently, a blueprint has no knowledge of the current status of its underlying resources. This RFC proposes how resources can report (and categorize) this information strictly for informational purposes.
- What use cases does it support?
  - This feature will enable developers to learn about the progression through the blueprint by looking at the owner status alone. It will also surface information about the health of the underlying resources directly on the owner.
  - This will also enable integrations that may be looking to create a UI around Cartographer, allowing the UI to display the health of the blueprint without having to make an API call to k8s to retrieve the status of every resource, and without having to know how to define health for every stamped object type.
- What is the expected outcome?
  - Users will have more insight into the progression of the blueprint and the health of each underlying resource.

# What it is
[what-it-is]: #what-it-is

This RFC proposes adding a field to the spec of all templates, `healthyConditionRule`. Template authors use field to define the conditions or fields on a stamped resource that determine whether the resource is healthy.

Cartographer will attempt to read the specified conditions and/or fields and reflect this state on the owner status. If Cartographer cannot determine health, it will show the status as unknown.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:
  urlPath: .status.artifact.url
  revisionPath: .status.artifact.revision

  healthyConditionRule:
    singleConditionType: Ready

  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository

---
apiVersion: carto.run/v1alpha1
kind: Workload
status:
  resources:
    - name: source-provider
      conditions:
        - type: Healthy
          status: 'True'
          message: 'Fetched revision: main/...'
          reason: ReadyCondition
          lastTransitionTime: "2022-03-23T13:08:22Z"
    - name: source-provider-2
      conditions:
        - type: Healthy
          status: 'Unknown'
```

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: deploy
spec:
  healthyConditionRule:
    multiMatch:
      healthy:
        matchConditions:
          - type: ReconcileSucceeded
            status: 'True'
      unhealthy:
        matchFields:
          - key: 'status.conditions[?(@.type=="ReconcileFailed")].status'
            operator: 'In'
            values: ['True']
            messagePath: '.status.usefulErrorMessage'

  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App

---
apiVersion: carto.run/v1alpha1
kind: Workload
status:
  resources:
    - name: deployer
      conditions:
        - type: Healthy
          status: 'False'
          message: "status.conditions['ReconcileFailed'].status: True: Some kapp useful error message"
          reason: MatchedField
          lastTransitionTime: "2022-03-23T13:08:22Z"
```

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: servicer
spec:
  healthyConditionRule:
    multiMatch:
      healthy:
        matchFields:
          - key: '.status.loadBalancer'
            operator: 'Exists'
      unhealthy:
        matchFields:
          - key: '.status.loadBalancer'
            operator: 'DoesNotExist'

  template:
    apiVersion: v1
    kind: Service

---
apiVersion: carto.run/v1alpha1
kind: Workload
status:
  resources:
    - name: service-creator
      conditions:
        - type: Healthy
          status: 'True'
          message: "status.loadBalancer: {}"
          reason: MatchedField
          lastTransitionTime: "2022-03-23T13:08:22Z"
```

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: configmap-creator
spec:
  healthyConditionRule:
    alwaysHealthy: {}
    
  template:
    apiVersion: v1
    kind: ConfigMap

---
apiVersion: carto.run/v1alpha1
kind: Workload
status:
  resources:
    - name: configmap-creator
      conditions:
        - type: Healthy
          status: 'True'
          reason: AlwaysHealthy
          lastTransitionTime: "2022-03-23T13:08:22Z"
```

# How it Works
[how-it-works]: #how-it-works

A stamped resource can be in one of three states:
1. 'Healthy'
2. 'Unhealthy'
3. 'Unknown'

It is up to each template author to determine what "healthy" and "unhealthy" mean for the given resource the template is stamping out. "Unknown" strictly represents that Cartographer has not been able to determine health.

Each template will now have a new field in the spec `healthyConditionRule` where authors can specify one of three ways to determine the health of the underlying resource for that template. If no `healthyConditionRule` is defined, Cartographer will default to listing the resource as `Healthy` once it has been successfully applied to the cluster and any relevant outputs have been read off the resource.

The `healthyConditionRule` requires defining one of three fields:
1. `alwaysHealthy: {}`
    * If defined, the resource will always be listed as healthy.
2. `singleConditionType: <condition type>`
    * If defined, the health of the resource will be determined by the single condition specified
      * e.g. if it is set to `singleConditionType: Ready`, the resource will be 'Healthy' if `Ready: True` and 'Unhealthy' if `Ready: False`. Otherwise, it will be 'Unknown'.
3. `multiMatch`
    * If defined, matchers for both the 'Healthy' state and the 'Unhealthy' state must be defined.
    * The matchers in the 'Healthy' state are ANDed together, whereas the matchers in the 'Unhealthy' state are ORed.
      * That is, a resource is considered to be 'Healthy' only if _all of_ the matchers are fulfilled, but is considered to be 'Unhealthy' if _any one of_ the matchers are fulfilled.
    * The supported matchers are `matchConditions` and `matchFields`. Each is an array, and an author can define either or both of these matchers for a given state.
    * For `matchConditions`, authors need only define the `type` and `status` of the condition for Cartographer to read. In the owner status, the full condition will be reflected (with the `reason`, `message`, etc.)
    * For `matchFields`, the spec is based off of `matchFields` in blueprint resource options. Authors need to define the `key` and `operator`, and potentially the `values` (if the operator is `In` or `NotIn`). Authors can also optionally define a `messagePath`, which is a path in the resource where Cartographer can pull off a message to display in the owner status.
    
Each owner's status.resources will now include a `conditions` field. To make room for future extensibility, this will be an array of conditions, though this RFC only defines one condition to be included here. The condition will be of `type` 'Healthy' and will report a `status` of 'True', 'False', or 'Unknown'. The condition will also have a `reason` which will be based off the `healthyConditionRule` as follows:
    
* If `alwaysHealthy: {}`, `reason: AlwaysHealthy`.
* If `singleConditionType: X`, `reason: XCondition`.
* If `multiMatch: ...`, `reason: (MatchedCondition|MatchedField)`.

If the `healthyConditionRule` is not `alwaysHealthy`, the condition will also have a message. For `singleConditionType`, the message will be propagated from the specified condition. For `multiMatch`, the message will specify the value of the matched condition or field, and then display the message of the matched condition or the message at the path specified in `matchFields`.

In the case of `multiMatch`, if there is more than one matching condition or field that caused the resource to enter it's current state, the first one read will be propagated by Cartographer.

Each owner will display a new condition in the status called `ResourcesHealthy` which will only be updated to `True` when all resources in the blueprint are healthy. Cartographer will not update the top level state of an owner to `Ready: True` until `ResourcesHealthy` is `True`.

# Migration
[migration]: #migration

Given that Cartographer will default to `Healthy` once the resource is applied and outputs are read, this will not be a breaking change.

# Drawbacks
[drawbacks]: #drawbacks

Why should we *not* do this?
- We only allow for reflecting three states, though this design allows for creating more states in the future.
- This design does not allow authors to define the states themselves.

# Alternatives
[alternatives]: #alternatives

- What other designs have been considered?
  - We could create a separate CRD to define the healthy condition rules and reference them from the templates or have them select templates based on the underlying resource type.
    - Separating these rules in a separate CRD would force us to start reconciling templates to see if there is an associated healthy condition rule CR. 
    - This would also force template authors to ship the healthy condition CR along with each of their templates.
    - If CR selected templates, this could be a bit confusing because the type of the stamped resource may be templated as well. Therefore, a template could end up stamping out different types of resources and so Cartographer would only be able to inform users which healthy condition rule is used for a given template after an owner is applied, and it could then be reflected in the owner status.
      - It would also mean that templates stamping out the same underlying type would not be able to define different healthy rules, if they desired to.
- What is the impact of not doing this?
  - Users/UI creators will have to have in depth knowledge of the underlying resources to determine the health of the owner.

# Prior Art
[prior-art]: #prior-art

N/A

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

- Should the two states be `Healthy` and `Unhealthy`?
- Is it right for Cartographer to default to `Healthy`?
- Should the owner condition `ResourcesHealthy` have any possible `reason`s other than `HealthyConditionRule`?

# Spec. Changes
[spec-changes]: #spec-changes
- All templates will now have an optional `healthyConditionRule` which will take one of three fields:
```yaml
apiVersion: carto.run/v1alpha1
kind: Cluster[Config|Deployment|Image|Source]Template
spec:
  # validation:MinProperties:1
  # validation:MaxProperties:1
  healthyConditionRule:
    alwaysHealthy: {}
    singleConditionType: <string>
    multiMatch:
      healthy:
        matchConditions:
          - type: <string>
            status: <string>
        matchFields:
          - key: <string>
            operator: <[In|NotIn|Exists|DoesNotExist]>
            values: [ <string> ]
            messagePath: <string>
      unhealthy:
        matchConditions:
          - type: <string>
            status: <string>
        matchFields:
          - key: <string>
            operator: <[In|NotIn|Exists|DoesNotExist]>
            values: [ <string> ]
            messagePath: <string>
```
- Owner will have a new condition of type `ResourcesHealthy`. Additionally, `status.resources` will now have `conditions` on each resource with a `Healthy` condition:
```yaml
apiVersion: carto.run/v1alpha1
kind:  [Workload|Deliverable]
status:
  conditions:
    - type: ResourcesHealthy
      status: <[True|False|Unknown]>
      reason: HealthyConditionRule
      message: <string>
      lastTransitionTime: metav1.Time 
  resources:
    - name: resource1
      conditions:
        - type: Healthy
          status: <[True|False|Unknown]>
          reason: <[AlwaysHealthy|<ConditionType>Condition|MatchedCondition|MatchedField]>
          message: <string>
          lastTransitionTime: metav1.Time
```
