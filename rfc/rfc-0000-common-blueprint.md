# Meta
[meta]: #meta
- Name: Common Blueprint Architecture
- Start Date: 2022-03-29
- Author(s): @emmjohnson @martyspiewak
- Status: Draft
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

This RFC proposes collapsing all current "Blueprints" and templates (not including `ClusterRunTemplate`, as that resource is in the domain of [Runnable](https://cartographer.sh/docs/v0.3.0/reference/runnable/)) into a common type, `ClusterBlueprint`. 
As part of this, the RFC proposes that outputs be dynamically typed via a new CR `ClusterBlueprintType` and remove the statically typed outputs that exist today (url, revision, image, config). This allows us to maintain the contract of swappable templates, without restricting outputs to a predefined set. 
Additionally, the selectors will be removed from blueprints and instead be specified in a new CR `ClusterBlueprintSelector`.


# Motivation
[motivation]: #motivation

- Why should we do this?

  - Today we have seven CRDs that are needed to create a supply chain and delivery. To simplify the data model for users, we can
collapse all seven CRDs into a few common CRDs. Additionally, whereas today the overhead to start reconciling templates would be adding five identical
reconcilers, this would enable us to have one reconciler for all blueprints and more easily add meaningful data to their statuses.

- What use cases does it support?

  - This would allow for snippets and users to specify any output they want, while maintaining strong typing.
  - The `ClusterBlueprintSelector` will enable users to separate RBAC between template authors and operators who want to restrict what blueprints can be selected.

- What is the expected outcome?

  - Simplified on-boarding and user experience.

# What it is
[what-it-is]: #what-it-is
* `ClusterBlueprints` will encapsulate everything the current templates allow, while adding `spec.resources` to enable defining multiple resources (i.e. snippets).
* A `ClusterBlueprint` can take one of `spec.resources`, `spec.template`, and `spec.ytt`. It will allow inlining of templates in `spec.resources`.
* A `ClusterBlueprint` must specify a `blueprintTypeRef` which defines the outputs required from the blueprint. The `ClusterBlueprint`'s `spec.outputs` must match those outputs.
* A `ClusterBlueprintType` defines only `spec.outputs`, which is an array of output names. By leaving the `name` key in the array, we leave open the possibility to later add `type` to each output, though this RFC currently proposes only supporting string outputs.
* A `ClusterBlueprintSelector` defines a `spec.blueprintRef`, `spec.selector` (per the v1alpha2 syntax), and `spec.ownerType` to define the type of owner to match against. This type takes an `apiVersion` and `kind` to leave open the possibility to later support other owner types, though this RFC proposes only supporting the existing workload and deliverable types.

Example with resources, inlining, two outputs, and workload selector:
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprint
metadata:
  name: my-blueprint
spec:
  blueprintTypeRef:
    name: my-baz-fizz-type
  
  # mutually exclusive - resources, template, ytt 
  resources:
    - name: first-resource
      inputs:
        - resource: zero-resource
          output: zoo
          name: zoo-resource-name
      inline:
        # mutually exclusive - template, ytt
        template: {}
        ytt: {}

        outputs:
          - name: foo
            valuePath: .status.url

    - name: second-resource
      # mutually exclusive - blueprintName, options
      options:
        - name: another-blueprint
        - name: diff-blueprint
      inputs:
        - resource: first-resource
          output: foo
          name: a-new-resource-name
        - resource: zero-resource
          output: zoo
          name: a-new-resource-name

    - name: third-resource
      # mutually exclusive - blueprintName, options
      blueprintName: another-another-blueprint
      inputs:
        - resource: first-resource
          output: foo
          name: a-new-resource-name
        - resource: second-resource
          output: bar
          name: another-new-name
  
  outputs:
    - name: baz
      valuePath: resources[first-resource].outputs[foo]
    - name: fizz
      valuePath: resources[second-resource].outputs[bar]

---
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprintType
metadata:
  name: my-baz-fizz-type
spec:
  outputs:
    - name: baz
    - name: fizz
  
---
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprintSelector
metadata:
  name: my-selector
spec:
  blueprintRef:
    name: my-blueprint
  ownerType:
    apiVersion: carto.run/v1alpha1
    kind: Workload
  selector:
    matchLabels: []
    matchFields: []
    matchExpressions: []
```

Example with template, one output, and deliverable selector:
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprint
metadata:
  name: my-blueprint-2
spec:
  blueprintTypeRef:
    name: my-url-type

  # mutually exclusive - resources, template, ytt (i.e. a blueprint is either a template or a supply chain)
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test
    data:
      foo: bar

  outputs:
    - name: url
      valuePath: .data.foo

---
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprintType
metadata:
  name: my-url-type
spec:
  outputs:
    - name: url

---
apiVersion: carto.run/v1alpha1
kind: ClusterBlueprintSelector
metadata:
  name: my-selector-2
spec:
  blueprintRef:
    name: my-blueprint-2
  ownerType:
    apiVersion: carto.run/v1alpha1
    kind: Deliverable
  selector:
    matchLabels: []
    matchFields: []
    matchExpressions: []
```

# How it Works
[how-it-works]: #how-it-works

This RFC does not propose any new functionality (other than snippets), but rather proposes re-defining the API and CR's to encapsulate the same functionality that exists today.


# Migration
[migration]: #migration

This RFC proposes all new CRDs, so there is no need for migration.

# Drawbacks
[drawbacks]: #drawbacks

Why should we *not* do this?

- We would need to support both this new architecture, as well as the currently architecture and CRDs for some period of time, which might be challenging.

# Alternatives
[alternatives]: #alternatives

- What other designs have been considered?
  - Typed blueprints (`ClusterSourceBlueprint`, `ClusterImageBlueprint`, etc.)

- Why is this proposal the best?
  - It allows users to define the blueprint types and the outputs they care about.

- What is the impact of not doing this?
  - We will not be able to reconcile templates easily.
  - Users will continue to struggle with the learning curve.

# Prior Art
[prior-art]: #prior-art

[knative duck types](https://pkg.go.dev/knative.dev/pkg/apis/duck#:~:text=Knative%20leverages%20duck%2Dtyping%20to,Addressable%20%2C%20Binding%20%2C%20and%20Source%20.) 

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

N/A

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

This RFC proposes adding new CRDs, but not changing the existing ones.
