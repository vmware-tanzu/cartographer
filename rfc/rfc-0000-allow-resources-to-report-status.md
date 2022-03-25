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

This RFC proposes creating a new CRD, `ClusterResourceStatusRule`. Template authors use this CRD to define the conditions or fields on a stamped resource that determine whether the resource is succeeding or failing. The template author then references this `ClusterResourceStatusRule` from their template.

Cartographer will attempt to read the specified conditions and/or fields and reflect this state on the owner status. If Cartographer cannot determine success or failure, it will show the state as unknown.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterResourceStatusRule
metadata:
  name: ready-rule
spec:
  healthy:
    matchConditions:
      - type: Ready
        status: 'True'
  unhealthy:
    matchConditions:
      - type: Ready
        status: 'False'

---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: source
spec:
  urlPath: .status.artifact.url
  revisionPath: .status.artifact.revision

  resourceStatusRef:
    name: ready-rule

  template:
    apiVersion: source.toolkit.fluxcd.io/v1beta1
    kind: GitRepository
    metadata:
      name: $(workload.metadata.name)$
    spec:
      ...
    
---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  ...
spec:
  ...
status:
  conditions:
    ...
  resources:
    - name: source-provider
      outputs:
        ...
      stampedRef:
        ...
      templateRef:
        apiVersion: carto.run/v1alpha1
        kind: ClusterSourceTemplate
        name: source
      stampedStatus:
        status: 'Healthy'
        conditions:
          - type: Ready
            status: 'True'
            message: 'Fetched revision: main/...'
            reason: GitOperationSucceeded
            lastTransitionTime: "2022-03-23T13:08:22Z"
    - name: source-provider-2
      outputs:
        ...
      stampedRef:
        ...
      templateRef:
        apiVersion: carto.run/v1alpha1
        kind: ClusterSourceTemplate
        name: source
      stampedStatus:
        status: 'Unknown'
```

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterResourceStatusRule
metadata:
  name: kapp-rule
spec:
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

---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: deploy
spec:
  resourceStatusRef:
    name: kapp-rule

  template:
    apiVersion: kappctrl.k14s.io/v1alpha1
    kind: App
    metadata:
      name: $(workload.metadata.name)$
    spec:
      ...

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  ...
spec:
  ...
status:
  conditions:
    ...
  resources:
    - name: deployer
      outputs:
        ...
      stampedRef:
        ...
      templateRef:
        apiVersion: carto.run/v1alpha1
        kind: ClusterTemplate
        name: deploy
      stampedStatus:
        status: 'Unhealthy'
        fields:
          - key: status.conditions['ReconcileFailed'].status
            value: True
            message: "Some kapp useful error message"
```

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterResourceStatusRule
metadata:
  name: service-rule
spec:
  healthy:
    matchFields:
      - key: '.status.loadBalancer'
        operator: 'Exists'
  unhealthy:
    matchFields:
      - key: '.status.loadBalancer'
        operator: 'DoesNotExist'

---
apiVersion: carto.run/v1alpha1
kind: ClusterTemplate
metadata:
  name: servicer
spec:
  resourceStatusRef:
    name: service-rule

  template:
    apiVersion: v1
    kind: Service
    metadata:
      name: $(workload.metadata.name)$
    spec:
      ...

---
apiVersion: carto.run/v1alpha1
kind: Workload
metadata:
  ...
spec:
  ...
status:
  conditions:
    ...
  resources:
    - name: service-creator
      outputs:
        ...
      stampedRef:
        ...
      templateRef:
        apiVersion: carto.run/v1alpha1
        kind: ClusterTemplate
        name: servicer
      stampedStatus:
        status: 'Healthy'
        fields:
          - key: status.loadBalancer
            value: {}
```

# How it Works
[how-it-works]: #how-it-works

A stamped resource can be in one of three states:
1. 'Healthy'
2. 'Unhealthy'
3. 'Unknown'

It is up to each template author to determine what "success" and "failure" mean for the given resource the template is stamping out. "Unknown" strictly represents that Cartographer has not be able to determine success or failure.

A `ClusterResourceStatusRule` requires defining matchers for both the `Healthy` state and the `Unhealthy` state.
The matchers in the `Healthy` state are ANDed together, whereas the matchers in the `Unhealthy` state are ORed. That is, a resource is considered to be `Healthy` only if _all of_ the matchers are fulfilled, but is considered to be `Unhealthy` if _any one of_ the matchers are fulfilled.

The supported matchers are `matchConditions` and `matchFields`. Each is an array, and an author can define either or both of these matchers for a given state.

For `matchConditions`, authors need only define the `type` and `status` of the condition for Cartographer to read. In the owner status, the full condition will be reflected (with the `reason`, `message`, etc.)
For `matchFields`, the spec is based off of `matchFields` in blueprint resource options. Authors need to define the `key` and `operator`, and potentially the `values` (if the operator is `In` or `NotIn`). Authors can also optionally define a `messagePath`, which is a path in the resource where Cartographer can pull off a message to display in the owner status.

Each template will now have a new field in the spec `resourceStatusRuleRef` where authors can specify the `ClusterResourceStatusRule` to use for that template. If no `ClusterResourceStatusRule` is defined, Cartographer will default to listing the resource as `Healthy` once it has been successfully applied to the cluster and any relevant outputs have been read off the resource.

Templates will now also have a status with a `Ready` condition which will be `True` if the referenced `ClusterResourceStatusRule` exists (or if no rule is referenced), and `False` otherwise. Blueprints will now wait to report that they are `Ready` until their referenced templates are also `Ready`.

Each owner will display a new condition in the status called `ResourcesSuceeded` which will only be updated to `True` when all resources in the blueprint be healthy. Cartographer will not update the top level state of an owner to `Ready: True` until `ResourcesHealthy` is `True`.

# Migration
[migration]: #migration

Given that Cartographer will default to `Healthy` once the resource is applied and outputs are read, this will not be a breaking change.

# Drawbacks
[drawbacks]: #drawbacks

Why should we *not* do this?
- The owner status will potentially grow significantly, based on how many matchers are in a `ClusterResourceStatusRule`.
- We only allow for reflecting three states, though this design allows for creating more states in the future.
- This design does not allow authors to define the states themselves.

# Alternatives
[alternatives]: #alternatives

- What other designs have been considered?
  - The `ClusterResourceStatusRule` could instead be inlined in the templates.
    - Separating these rules in a separate CRD allows for re-usability of those rules.
  - We could also not require the templates to specify the `ClusterResourceStatusRule` and instead have those rules automatically select templates based on the `kind` and `apiVersion` of the resource they are stamping out.
    - This could be a bit confusing because the type of the stamped resource may be templated as well. Therefore, a template could end up stamping out different types of resources and so Cartographer would only be able to inform users which `ClusterResourceStatusRule` is used for a given template after an owner is applied, and it could then be reflected in the owner status.
- What is the impact of not doing this?
  - Users/UI creators will have to have in depth knowledge of the underlying resources to determine the health of the owner.

# Prior Art
[prior-art]: #prior-art

N/A

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

- Should the two states be `Healthy` and `Unhealthy`?
- Is it right for Cartographer to default to `Healthy`?

# Spec. Changes
[spec-changes]: #spec-changes
- All templates will now have an optional `resourceStatusRuleRef` which will take a `name`. In the future, it could also take a type and namespace if we introduce namespace-scoped resourceStatusRules.
- Owner `status.resources` will now have `stampedStatus` on each resource:
```yaml
status:
  resources:
    - name: resource1
      ...
      stampedStatus:
        status: <[Healthy|Unhealthy|Unknown]>
        conditions:
          - <metav1.Condition>
        fields:
          - key: <string>
            value: <string>
            message: <string>, omitempty
```
- A new CRD, `ClusterResourceStatusRule` will be created with the following spec:
```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterResourceStatusRule
metadata: {}
spec:
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
