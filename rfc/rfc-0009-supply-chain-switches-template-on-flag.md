# Draft RFC <0009> Supply Chain Switches Template Dynamically

## Summary

Platform creators will need to handle multiple types of workloads. Cartographer
must support that behavior through either multiple supply chains, or with
supply chains that can substitute templates dynamically.

## Motivation

The Developer Productivity program auto-generates workloads from CLI commands.
Those workloads match with supplychains and templates that are written by the
dev-prod team as well. The team would like to limit the number of supplychains
that must be written and maintained. Small differences in workload can lead to
needing an entirely different supply chain. For example, what code source is used.

Further info on their use case can be read [here](https://docs.google.com/document/d/1TVqlNqTyCMlp_yNs9_F80QAIQSC5bxm05vSyWyeGMSw/edit).
See also this [discussion](https://vmware.slack.com/archives/C01UX69LJCB/p1629408443051200).

## Possible Solutions

Cartographer could make conditional choices based on labels on the workload.
Alternatively, Cartographer could consider all fields available in the templating context when
making conditional choices.

### Use label selectors

#### Selecting among multiple supply chains

Cartographer is already capable of matching a supply chain to a workload through selectors.
With no further effort, a platform can be created that has multiple supply chains with different
purposes or requirements. By specifying characteristics of workloads in the workload labels,
devs can provide information to operators necessary to choose the right supply chain for
each workload.

While this behavior is currently possible, relying only on this selection would lead to a platform
with a large number of supply chains.

Workarounds, for example templates that use ytt templating to embed conditional behavior,
would represent a code smell that this primitive is not adequate for Cartographer consumers.

#### Selecting among multiple templates at a given step in a supply chain

A supply chain resource could reference a template based on the presence of a
label in the workload. Below is a possible example:

```yaml
apiVersion: kontinue.io/v1alpha1
kind: ClusterSupplyChain
metadata:
  name: responsible-ops
spec:
  selector:
    app: web
  resources:
    - name: source-provider
      templateRefs:                               # <--- now a list
      - kind: ClusterSourceTemplate
        name: git-template
        selector:
          app.tanzu.vmware.com/source-type: git    # <--- label whose presence will trigger use of this template
      - kind: ClusterSourceTemplate
        name: imgpkg-bundle-template
        selector:
          app.tanzu.vmware.com/source-type: imgpkg-bundle # <--- alternate label for alternate template

---
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    app: web
    app.tanzu.vmware.com/source-type: git # <--- selector on the workload
spec:
  source:
    git:
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main
```

This approach allows far fewer (potentially just one) supply chain to under gird an app
platform.

One shortfall of this approach is that the app-dev must explicitly provide all information
about the workload when submitting the workload object. This assumes that the app operator
has been able to communicate all the requisite information required.

This concern (the need for the dev to be aware of all information necessary to provide) could
be alleviated through auto-labelling. This could be achieved with a mutating webhook on
workloads, which could apply a label to a workload based on characteristics of the workload.
Such an approach would work for all characteristics observable from the workload, but would
likely exclude code characteristics (for example, language of the app).

The label approach also assumes that all relevant information is knowable at workload submission
time. But it is conceivable that transformations that take place during the supply chain (for example,
application of a convention), could have conditional impact on later supply chain steps.

### Condition on any/all information available in the templating context

When a supply chain arrives at a resource/step, it provides a templating context to the template.
Therein the template can access all artifacts provided from previous steps, params defined and the
workload object itself (including the spec and the metadata). Allowing the supply chain to leverage
this field would add additional power to the conditional.

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
      templateRefs:                               # <--- again, a list
      - kind: ClusterSourceTemplate
        name: git-template
        requisite:
          path: workload.spec.source.git        # <--- path to field in template context
          value: *                              # <--- value that will trigger use of this template, in this case any value
      - kind: ClusterSourceTemplate
        name: imgpkg-bundle-template
        requisite:
          path: workload.spec.source.image      # <--- Alternate value for alternate conditional
          value: *

---
apiVersion: kontinue.io/v1alpha1
kind: Workload
metadata:
  name: petclinic
  labels:
    app: web
spec:
  source:
    git:                                        # <--- Presence of this field will determine the conditional
      url: https://github.com/spring-projects/spring-petclinic.git
      ref:
        branch: main
```

1. As labels are present in the workload metadata, any behavior possible using just label selectors
  will be possible using the entire templating context.
2. As the templating context is updated during the processing of the supply chain, the context
  can support conditionals that rely on information emergent during the supply chain processing.
3. As the templating context has access to all of the information that would be available to a
  mutating webhook, it obviates the need for such a webhook. Rather than specify "If workload spec
  has characteristic X, apply label Y" and later "If label Y is present, do Z", the template context
  approach can simply state "If workload spec has characteristic X, do Z".
