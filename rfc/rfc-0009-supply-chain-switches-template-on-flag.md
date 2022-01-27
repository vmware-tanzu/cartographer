# RFC 0009 Supply Chain Switches Templates

## Summary

Platform creators will need to handle multiple types of workloads. While Cartographer
supports that behavior currently through allowing multiple supply chains, it should
reduce the number of supply chains necessary by allowing individual supply chains
to choose between templates within a step.

## Motivation

A team creating a developer platform on top of Cartographer reached out because
they maintain supply chains for their platform. They have found that there are
many small differences in workloads that necessitate making a change to one
template or another. Currently, such a difference can be handled by:

1. Using ytt within one template to define two different objects.
2. Creating two templates for to handle the difference and then two supply chains
   that are the same identical except for one step that uses template A or B.

The team has expressed that they prefer having separate templates (as opposed to
maintaining ytt templates) but is concerned about the number of supply chains that
will be needed to pursue this approach. Their hope is to keep the number of
supply chains small, even as many small differences between workloads could lead to
a combinatoric explosion of paths to prod.

## Explanation

### Current state of Cartographer

Cartographer is already capable of matching a supply chain to a workload through selectors.
With no further effort, a platform can be created that has multiple supply chains with different
purposes or requirements. By specifying characteristics of workloads in the workload labels,
devs can provide information to operators necessary to choose the right supply chain for
each workload.

While this behavior is currently possible, relying only on this selection would lead to a platform
with a large number of supply chains.

Workarounds, for example templates that use ytt templating to embed conditional behavior,
would represent a code smell that this primitive is not adequate for Cartographer users.

### Condition template choice on information available from the workload

When creating objects of the supply chain, Cartographer has the workload definition available.
Allowing the supply chain to leverage the workload would add additional power to the conditional.

Each step/resource in a supply chain will continue to define a templateRef, which will continue to
specify a single template kind (e.g. ClusterImageTemplate). Rather than specifying the name of a
particular template, authors can provide `options` which is a list of options. An option is
a `name` and a `selector`. The only selector adopted by this RFC is `MatchFields` (though other
selectors can be added later). A MatchFields contains:
- `key`: a path on the workload object. The path must be prefixed by `workload`.
- `operator`: One of `In`, `NotIn`, `Exists`, `DoesNotExist`
- `values`: Required and allowed only when the operator is either `In` or `NotIn`. A list of json values.
  (e.g. both strings and integers are valid values)

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
      templateRef:
        kind: ClusterSourceTemplate
        options:                                  # <--- a list
        - name: git-template
          selector:
            matchFields:
              - key: workload.spec.source.git       # <--- path to field in template context
                operator: Exists
        - name: imgpkg-bundle-template
          selector:
            matchFields:
              - key: workload.spec.source.image       # <--- path to field in template context
                operator: Exists

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

### Error conditions

Exactly one template should be chosen at each step. If no template is chosen, or multiple templates match, an error
should be thrown.

### Matching syntax

Kubernetes resources "such as Job, Deployment, ReplicaSet, and DaemonSet, support set-based requirements". The
syntax for matchExpressions allows for asserting `In` or `NotIn` a specified array of values, or to assert
that a key `Exists` or `DoesNotExist`. matchExpressions is a list of assertions, which are ANDed.
[link](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#resources-that-support-set-based-requirements)

## Notes

1. While this RFC has referred to workloads and supply chains throughout, the proposal is to apply this
  behavior to Delivery/Deliverable as well.
2. This RFC benefited from comments in the pull request, discussions of the Cartographer team and community
  and ideas in [this proposal](https://gist.github.com/squeedee/723be000c4f2ee40ce4c9ac020cbf4fc) 
  from [Rasheed Abdul-Aziz] (https://github.com/squeedee)

## Concerns

Deterministic supply chains are desirable. That is, given a workload it is valuable to be able to
state which supply chain will be chosen and what the path through that supply chain will be.

However, it is not certain that Cartographer will have the luxury of determinism. It is possible that the
output of a choreographed resource may be important to the next step of a supply chain. Essentially, what if
an object emits some values that can be handled by controller A, and other objects that can only be handled
by controller B? There would need to be a step which could switch on the emitted values. This RFC does not
handle this case. An RFC to use the entire templating context for template choice would address this use case.

## Alternatives

### Label selectors for choosing templates

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

This approach similarly allows far fewer (potentially just one) supply chain to under gird an app
platform.

One shortfall of this approach is that the app-dev must explicitly provide all information
about the workload when submitting the workload object. This assumes that the app operator
has been able to communicate all the requisite information required.

This concern (the need for the dev to be aware of all information necessary to provide) could
be alleviated through auto-labelling. This could be achieved with a mutating webhook on
workloads, which could apply a label to a workload based on characteristics of the workload.
Such an approach would work for all characteristics observable from the workload, but would
likely exclude code characteristics (for example, language of the app).

However, as the workload has access to all of the information that would be available to a
mutating webhook, the approach recommended in this RFC obviates the need for such a webhook.
Rather than specify "If workload spec has characteristic X, apply label Y" and later "If
label Y is present, do Z", the whole workload approach can simply state "If workload spec
has characteristic X, do Z".

The label approach also assumes that all relevant information is knowable at workload submission
time. But as mentioned in `Concerns`, it is conceivable that in the future there will be
outputs in the supply chain that could have conditional impact on later supply chain steps. It
is not clear that the label approach could be leveraged to handle that case. That is choosing the
label approach introduces more risk of rewriting in the future.
