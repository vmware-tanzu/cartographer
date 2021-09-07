# Draft RFC <0009> Supply Chain Switches Template Dynamically

## Summary

The Developer Productivity team would like to be able to write 1 supply chain
which can dynamically use Template A or Template B in a component depending
on the state of the workload. This RFC proposes methods for achieving that
outcome.

## Motivation

The Developer Productivity program auto-generates workloads from CLI commands.
Those workloads match with supplychains and templates that are written by the
dev-prod team as well. The team would liek to limit the number of supplychains
that must be written and maintained. Small differences in workload can lead to
needing an entirely different supply chain. For example, what code source is used.

Further info on their use case can be read [here](https://docs.google.com/document/d/1TVqlNqTyCMlp_yNs9_F80QAIQSC5bxm05vSyWyeGMSw/edit).
See also this [discussion](https://vmware.slack.com/archives/C01UX69LJCB/p1629408443051200).

## Possible Solutions

Kontinue could choose which template to use based on a flag specified in the workload.
Alternatively, Kontinue could attempt to deploy the default template and if the necessary
values are not found, fall back to attempt a different template.

### Switch on Flag
A supply chain component could reference a template based on the presence of a
flag in the workload. Below is a proposed example using the keyword `requisite`:

```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops
spec:
  selector:
    app: web
  components:
    - name: source-provider
      templateRefs:                               # <--- now a list
      - kind: ClusterSourceTemplate
        name: git-template
        requisite:
          path: workload.spec.flags.source-type    # <--- path to field in workload
          value: git                              # <--- value that will trigger use of this template
      - kind: ClusterSourceTemplate
        name: imgpkg-bundle-template
        requisite:
          path: workload.spec.flags.source-type    # <--- ideally path is same for all templateRefs in component
          value: imgpkg-bundle

---
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    app: web
spec:
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main
  flags:                                           # <--- map[string]string
    source-type: git
```

### Switch through fall-back strategy
Kontinue could allow users to define fall-back values if a given value does not work.
This could happen in 2 different ways:
- Allow a supply-chain component to define a fallback template if the first creation attempt fails
- Allow a template to define a fallback value if the first templated value does not resolve

#### Supply Chain defines fall-back template
Replace the single template ref for each component with a list of template refs. Kontinue would try them in order and when one successfully deploys, it would proceed to creating the next component, skipping the other template refs for the current component. An example definition:

```
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops
spec:
  selector:
    ...
  components:
    - name: source-provider
      templateRefs:
      - kind: ClusterSourceTemplate
        name: git-resource-template
      - kind: ClusterSourceTemplate
        name: imgpkg-template
      - kind: ClusterSourceTemplate
        name: image-template
```

#### Template defines a fallback value
a BuildTemplate might have some definition:

```
apiVersion: kontinue.io/v1alpha1
kind: ClusterBuildTemplate
metadata:
  name: example-build
spec:
  params:
    - name: url
      default: some-url
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example-build-configmap
    data:
      templated-field: $(sources[0].url|params[?(@.name=="url")].value)$
  imagePath: ...
```

We can see the introduction of the `|` character to separate possible paths to use to fill this field.

## Cross References and Prior Art

{{Reference other similar implementations, and resources you are using to draw inspiration from}}
