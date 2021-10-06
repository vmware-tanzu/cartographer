# Draft RFC <0010> Supply Chains Supply Default Params

## Summary

Template authors are empowered to provide default params in the case that
supply chain authors do not provide them. Cartographer does not use defaults
in the case that supply chain authors expected workloads to provide a param and
none is present. Possible solutions are explored.

## Motivation

1. A supply chain author may have a reasonable default in mind for a param. They may
wish to allow workloads to override this value. At present, cartographer cannot
handle this case.
2. The current approach to params requires specifying a path on the workload in
  supply chain components. This approach has led to a recursive templating approach
  which is complicated. This recursive approach cannot be carried out by ytt.

## Detailed Explanation

### Name matching
The supply-chain will use the name-matching pattern that the templates use.
In the same way that the template assumes that the supply chain will have
a param with a matching name, the supply chain will assume the workload will have a matching
name param.

This would remove the only instance where a field in the supply chain is templated.
```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
spec:
  params:
    - name: some-param             # <--- currently templates do not use jsonpath to specify supplychain location
      default: "a fantastic value"
template:
  apiVersion: carto.run/v1alpha1
  kind: Deliverable
  value:
    first: $(params[?(@.name=="some-param")].value)$

---
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
  components:
  - name: some-component
    params:
      - name: some-param         # <--- proposing that supply-chains similarly should not use jsonpath to specify workload location
        default: "a great value"
```

#### Rationale
As noted in the motivation the introduction of jsonpath in the supply chain has increased
complication of the controller. This is not necessary, we already have the pattern of matching params by name only.

### Supply Chain Author Control
In order to empower the app operator (the supply chain author) to control what
fields can be over-written and which cannot, each param field would have 2 sub-fields alongside name:
value and default. These would be mutually exclusive, the supply chain could not specify both for a param.
For each param, the supply chain author could:
1. Specify neither default nor value. In which case the default value in the template would apply.
A matching param value specified by the workload would be ignored.
This would be equivalent to the workload not specifying the param at all.
2. Specify the 'value' field. In which case that value would override the template's default.
A matching param value specified by the workload would be ignored.
3. Specify a 'default' field. If the workload specifies a matching param,
that value would override the supply chain and the template default.
If the workload does not specify a matching param, the supply chain default
would override the template value.

#### Rationale
The app operator is the persona that should have final call on what objects are
stamped out in a supply chain. The above structure allows the app operator to declare
a parameter value, to delegate to the template or to delegate to the dev (workload author).

### Avoiding name collisions
In the section 'name matching', we specified "the supply chain will assume the workload will have a matching
name param". Here we augment/change that match. The workload should have a param whose name is:
"<name_of_the_supply_chain_component>/<name_of_the_param>". E.g. for this supply-chain:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
  components:
  - name: some-component
    params:
      - name: some-param
        default: "a great value"
  - name: another-component
    params:
       - name: some-param
         default: "a greater value"
```

the workload would be expected to have params named thusly:

```yaml
apiVersion: carto.run/v1alpha1
kind: Workload
spec:
  params:
    - name: some-component/some-param
      value: "a small value"
    - name: another-component/some-param
      value: "a smaller value"
```

#### Rationale
Template authors may define the same param name, but need different values.
There must be a way to disambiguate these.

Example: Imagine two templates that ask for a parameter like 'disk-space'.
One is using 'disk-space' to know how much space to allocate for a test,
another is passing that value in to the image being built. (A hand-wavy
example, that nonetheless illustrates the problem)

Those two templates should not have to have knowledge of each other, but the
name collision must still be disambiguated. The supply chain is the mediator
between templates and workloads. So the supply chain must have some way of
passing the right 'disk-space' values from the workload to template A and
template B. (Note that in our current architecture, this problem could occur
and would require the supply chain author to enforce that workloads specified
two differently named fields in workload.spec.params)

## Alternatives

### Alternatives to name matching

1. The supply chain can provide a default value by continuing to use jsonpath
  templates to identify the expected value for the workload. e.g.
      ```yaml
      apiVersion: carto.run/v1alpha1
      kind: ClusterSupplyChain
      spec:
        components:
        - name: some-component
          templateRef:
            ...
          params:
            - name: some-param
              value: $(workload.spec.params[?(@.name=="some-param")].value)$
              default: "a great value"  # <--- new field
        ```

2. The template could substitute its default parameter if the supplychain fails
   to find an expected value in the workload. This would not require any changes to the api.
   At the same time, I believe this does not preserve the separation of concerns.
   The author of a template and the author of a supply chain may be different. Their
   different contexts may lead them to different choices of defaults. Cartographer
   should empower both authors.

3. Drop all interpolation in the SupplyChain itself (i.e., only support constant params)
   Allow SupplyChain-provided params to be defaulted in templates (e.g., via supplyChainParams:)
   Separately, allow Workload-provided params to be defaulted in templates (e.g., via workloadParams:)

### Alternatives to avoiding name collisions
Supply chains add another field to params such as workload-param-name.
If that field is empty, we could assume that the workload param name in question
is the same as the param name. Example:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata: ...
spec:
  components:
    - name: image-provider
      templateRef: ...
      params:
        - name: disk-space
          default: 10GB
          workload-param-name: disk-space-for-images
```
