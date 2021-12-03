# RFC 0010 Supply Chains Supply Default Params

# Status
Accepted

## Summary

Template authors have always been empowered to provide default params in the case that
supply chain authors do not provide a parameter value. But Cartographer does not allow
supply chain authors to have similar delegation and fallback, e.g. to provide a default
value that can be replaced by a value in the workload. With this RFC, supply chain authors
can choose to provide a `default` or a `value` for each parameter. When providing a `default`
workload authors will be able to provide a value that is respected.

## Motivation

1. A supply chain author may have a reasonable default in mind for a param. They may
wish to allow workloads to override this value. At present, cartographer cannot
handle this case.
2. The current approach to params requires specifying a path on the workload in
  supply chain resources. This approach has led to a recursive templating approach
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
  resources:
  - name: some-resource
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

### Supply Chain and Resource Level Specification

Allow supply chain authors to set a parameter value for the entire supply chain.

E.g. in the following supply chain, any template could access the parameter `color`

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
    params:         # <=== apply to all resources in supply chain
      - name: color
        value: blue
    resources:
      ...
```

A supply chain author can override that choice for any individual resource.
E.g. in the following supply chain, the resource 'some-different-resource' would
supply the value `red` for the parameter `color`

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
    params:
      - name: color
        value: blue
    resources:
    - name: some-different-resource
      templateRef:
        ...
      params:
        - name: color  # <=== overrides the value supplied by the top level params
          value: red
```

The resource level specification clobbers the supply chain level specification.
This includes the choice of `default` vs `value`.

E.g. in this case, 'some-different-resource' would allow the workload to override.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
    params:
      - name: color
        value: blue  # <=== does not allow workload value to supersede
    resources:
    - name: some-different-resource
      templateRef:
        ...
      params:
        - name: color
          default: red  # <=== for this resource, the workload value will supersede
```

#### Rationale
While top level supply chain values may be useful, different resources may need to be
treated differently. Resource A and B may be able to use the same value for a param,
but resource C may need a different value. That value might be unknowable for the supply
chain author and require input from the workload author. This framework offers the
supply chain author full control of such delegation.

### Putting it all together

The order of precedence is therefore:

supply chain **value** (highest precedence)
workload **value**
supply chain **default**
template **default** (lowest precedence)

While supply chain is determined by the top level params and the resource level
params. If the resource level param is specified, the top level param is ignored.  

## Alternatives

### Alternatives to name matching

1. The supply chain can provide a default value by continuing to use jsonpath
  templates to identify the expected value for the workload. e.g.
      ```yaml
      apiVersion: carto.run/v1alpha1
      kind: ClusterSupplyChain
      spec:
        resources:
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

## Unhandled Corner Case: Avoiding name collisions
There is a concern about name collisions from different templates. Template authors may define
the same param name, but need different values. There must be a way to disambiguate these.

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

As noted in comments, **this RFC does not address this corner case.**
Some approaches that were discussed (and are not implemented):

### Supply chain points to different workload param name
Supply chains add another field to params such as workload-param-name.
If that field is empty, we could assume that the workload param name in question
is the same as the param name. This would be an additive (non-breaking) change.

Example:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
metadata: ...
spec:
  resources:
    - name: image-provider
      templateRef: ...
      params:
        - name: disk-space
          default: 10GB
          workload-param-name: disk-space-for-images
```

### Workload aware of component names
In the section 'name matching', we specified "the supply chain will assume the workload will have a matching
name param". Here we augment/change that match. The workload should have a param whose name is:
"<name_of_the_supply_chain_component>/<name_of_the_param>". E.g. for this supply-chain:

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSupplyChain
spec:
  resources:
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
