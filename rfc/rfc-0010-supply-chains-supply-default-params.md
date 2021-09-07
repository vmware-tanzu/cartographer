# Draft RFC <0010> Supply Chains Supply Default Params

## Summary

Template authors are empowered to provide default params in the case that
supply chain authors do not provide them. Cartographer does not use defaults
in the case that supply chain authors expected workloads to provide a param and
none is present. Possible solutions are explored.

## Motivation

A supply chain author may have a reasonable default in mind for a param. They may
wish to allow workloads to override this value. At present, cartographer cannot
handle this case.

## Possible Solutions

Problem: the supply chain provides a path to a workload param, but a workload
does not provide a value at that path.

1. The supply chain can provide a default value.
   1. The supply chain can continue to use jsonpath templates to identify the
      expected value for the workload. e.g.
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
   2. The supply-chain can use the name-matching pattern that the templates use.
      In the same way that the template assumes that the supply chain will have
      a param with a matching name (or else the default value should be
      substituted), the supply chain can assume the workload will have a matching
      name param.

      This would remove the only instance where a field in the supply chain is templated.
      ```yaml
      apiVersion: carto.run/v1alpha1
      kind: ClusterSourceTemplate
      spec:
        params:
          - name: some-param             # <--- templates do not use jsonpath to specify supplychain location
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
            - name: some-param         # <--- proposing dropping the 'value' field
              default: "a great value"
      ```

2. The template could substitute its default parameter if the supplychain fails
   to find an expected value in the workload. This would not require any changes to the api.
   At the same time, I believe this does not preserve the separation of concerns.
   The author of a template and the author of a supply chain may be different. Their
   different contexts may lead them to different choices of defaults. Cartographer
   should empower both authors.

## Cross References and Prior Art

{{Reference other similar implementations, and resources you are using to draw inspiration from}}
