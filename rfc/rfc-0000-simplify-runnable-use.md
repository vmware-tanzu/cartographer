# Meta
[meta]: #meta
- Name: Simplify Runnable Usage
- Start Date: 2022-03-08
- Author(s): [waciumawanjohi](https://github.com/waciumawanjohi)
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

Users will be able to create a Supply Chain Template and use a flag to
indicate that the object being stamped out is an immutable object. If they
do so, they will need to provide a success criteria. Given these inputs,
Cartographer will create a Runnable for the user and ensure that an
appropriate ClusterRunTemplate exists.

# Motivation
[motivation]: #motivation

Runnable exists in Cartographer to allow users to create and leverage
resources that are immutable, one time use objects. The prime example of
this is Tekton, which uses immutable pipelinerun and taskrun objects to
accomplish work. Users can leverage Tekton for testing and more, prime use
cases Cartographer supports.

While useful, Runnable represents the greatest complexity in Cartographer's
user interface. To use an immutable object in a supply chain users must
first template the object in a ClusterRunTemplate. Then they must template a
Runnable in one of the Cartographer template types used by supply
chains/deliveries (e.g. ClusterSourceTemplate, ClusterImageTemplate, etc).

The extra burden of this complication can observed in a number of ways:

1. In office hour conversations, the Cartographer team has been given
   feedback that the layering of objects presents a conceptual barrier for
   users. In that conversation it was suggested that Cartographer not add
   more such layers.
2. I've presented to teams of internal VMware engineers and VMware customers
   considering Tanzu Application Platform (for which Cartographer is a core
   component). Weeks later I've been asked, to clarify only a single part of
   Cartographer, why Tekton is embedded as part of Runnable.
3. In our examples we demonstrate how users can create an application
   platform from Cartographer. It takes only one example to demonstrate a
   simple supply chain that goes from source code to deployment. It takes
   two examples to demonstrate how to add testing to a supply chain, one
   example focused on using Runnables on their own and another example
   focused on using Runnable in the context of a supply chain.
4. In writing tutorials, the same pattern has been necessary: first to
   explain Runnable on its own and then to demonstrate how to use Runnable
   with Cartographer.

This complication is a needless barrier to entry. Cartographer should
be simple to use for our prime use cases, of which testing with Tekton is
one.

The pattern of using Runnable in a supply chain is predictable and
programmable. Rather than asking all app operators to learn this pattern,
Cartographer should create Runnables when supply chain templates indicate
their need.

# What it is
[what-it-is]: #what-it-is

Supply Chain/Delivery templates will add a field that will indicate that the
object templated needs to be wrapped by runnable. This will be an enum
field, with possible values `Mutable` and `Immutable`. The field will default to
mutable and users will not be required to specify it. The field will be
called `lifecycle`.

When a template specifies `lifecyle:immutable`, rather than creating the object
templated in the supply chain template, Cartographer will create a Runnable
and ensure that there is an appropriate ClusterRunTemplate. The
ClusterRunTemplate will template out all of the fields that vary and the
Runnable will provide all of these fields as inputs. The ClusterRunTemplate
will provide outputs that match the outputs of the supply-chain template in
question (ClusterSourceTemplate, ClusterImageTemplate, etc).

## Current example
This is an example of what an app operator would need to do today to use
Runnable in a supply chain (taken from
https://github.com/vmware-tanzu/cartographer/tree/main/examples/testing-sc):

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
   name: md-linting-pipelinerun
spec:
   template:
      apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
         generateName: $(runnable.metadata.name)$-pipeline-run-
      spec:
         pipelineRef:
            name: linter-pipeline
         params:
            - name: repository
              value: $(runnable.spec.inputs.repository)$
            - name: revision
              value: $(runnable.spec.inputs.revision)$
         workspaces:
            - name: shared-workspace
              volumeClaimTemplate:
                 spec:
                    accessModes:
                       - ReadWriteOnce
                    resources:
                       requests:
                          storage: 256Mi
   outputs:
      url: spec.params[?(@.name=="repository")].value
      revision: spec.params[?(@.name=="revision")].value

---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
   name: source-linter
spec:
   template:
      apiVersion: carto.run/v1alpha1
      kind: Runnable
      metadata:
         name: $(workload.metadata.name)$-linter
      spec:
         runTemplateRef:
            name: md-linting-pipelinerun
         inputs:
            repository: $(workload.spec.source.git.url)$
            revision: $(workload.spec.source.git.revision)$
         serviceAccountName: pipeline-run-management-sa
   urlPath: .status.outputs.url
   revisionPath: .status.outputs.revision
```

## Proposed example
Under this proposal, users would only specify one object with a new immutable
flag.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate    # <== Now a source template
metadata:
   name: source-linter
spec:
   template:
      apiVersion: tekton.dev/v1beta1
      kind: PipelineRun
      metadata:
         name: $(workload.metadata.name)$-linter      # <== Refers to workload
      spec:
         pipelineRef:
            name: linter-pipeline
         params:
            - name: repository
              value: $(workload.spec.source.git.url)$      # <== Refers to workload
            - name: revision
              value: $(workload.spec.source.git.revision)$ # <== Refers to workload
         workspaces:
            - name: shared-workspace
              volumeClaimTemplate:
                 spec:
                    accessModes:
                       - ReadWriteOnce
                    resources:
                       requests:
                          storage: 256Mi
   urlPath: spec.params[?(@.name=="repository")].value    # <== Refers to path on the immutable object that will be created
   revisionPath: spec.params[?(@.name=="revision")].value
   lifecycle: immutable                                         # <== Flag
```

# How it Works
[how-it-works]: #how-it-works

When Cartographer processes a template with the immutable flag, rather than
creating/updating the templated object it will create/update a Runnable and
ensure that a matching ClusterRunTemplate exists.

What is meant by "ensure that a matching ClusterRunTemplate exists"? For any
given supply chain template (a cluster scoped object) there will only need
to be one ClusterRunTemplate. All of the N Runnables created by N workloads
will reference this one ClusterRunTemplate. Only if the supply chain
template is updated will this ClusterRunTemplate object need updating.

## Inputs
The templated object will have two types of fields: hard-coded and templated.
Hard coded fields can simply be copied directly into the ClusterRunTemplate.
Templated fields will be translated. Rather than reading from the specified
location, these will read from the runnable's inputs field. This can be done
simply by prepending the template paths with runnable.spec.inputs

Example:
```yaml
kind: ClusterSourceTemplate    # <== Now a source template
spec:
   template:
         params:
            - name: repository
              value: $(workload.spec.source.git.url)$
```

Will translate to a ClusterRunTemplate with:
```yaml
kind: ClusterRunTemplate
spec:
   template:
         params:
            - name: repository
              value: $(runnable.spec.inputs.workload.spec.source.git.url)$
```

and a Runnable with
```yaml
kind: Runnable
spec:
   inputs:
      workload.spec.source.git.url: *some-value-from-workload.spec.source.git.url*
```

## Names
The runnable will be created with the same name assigned to the templated
object with the object kind appended.

Why append the object kind? We can imagine a supply chain that had one
template to create a tekton pipelinerun and another to create a tekton
taskrun. The template authors would not worry about name collisions, as they
are stamping out different objects. But when they specify that these are
immutable objects they may be surprised to see that two objects of the same
type are created, which now can have a name collision. To comply with
kubernetes name restrictions, slashes in the name must become dashes.

From the example above, assuming a workload with the name `app`:

```yaml
---
kind: Runnable
metadata:
   name: app-linter-pipelinerun-tekton.dev-v1beta1
```

The ClusterRunTemplate will be created with the same name as the supply
chain template, with the supply chain template type appended (e.g. appending
"-image" for a ClusterImageTemplate with the immutable flag present)

From the example above:

```yaml
---
kind: ClusterRunTemplate
metadata:
   name: source-linter-source
```

The Runnable will have a runTemplate ref that matches:
```yaml
---
kind: Runnable
spec:
   runTemplateRef:
      name: source-linter-source
```

## Outputs
The ClusterRunTemplate will specify outputs that match the output paths on
the template. Given the example above:

```yaml
---
kind: ClusterRunTemplate
spec:
   outputs:
      url: spec.params[?(@.name=="repository")].value
      revision: spec.params[?(@.name=="revision")].value
```

When reading the runnable, Cartographer will look at the fields that match
the type of template. E.g. given a ClusterSourceTemplate with a
`lifecycle:immutable` field, Cartographer will always create a Runnable and
always read the runnable's .status.outputs.url field and .status.outputs.
revision field. Similarly for image, carto will read .status.outputs.image;
for config it will read .status.outputs.config.

## Labels
The Runnable will be created with all the requisite labels of an object
created from a supply chain template:

```yaml
kind: Runnable
metadata:
  labels:
    carto.run/workload-name:
    carto.run/workload-namespace:
    carto.run/supply-chain-name:
    carto.run/resource-name:
    carto.run/template-kind:
    carto.run/cluster-template-name:
```

The ClusterRunTemplate will be created only with labels referring to the
template:

```yaml
kind: ClusterRunTemplate
metadata:
  labels:
    carto.run/template-kind:
    carto.run/cluster-template-name:
```

# Migration
[migration]: #migration

This is an additive change and does not present any breaking changes. Users
who wish to continue to hand create ClusterRunTemplates and template out
Runnables will be able to do so.

# Drawbacks
[drawbacks]: #drawbacks

- This RFC exchanges ease of usage for the app operator for complexity in the
controller. If we decide that app operators have no issue with creating
ClusterRunTemplates and templating Runnables, there is no need to undertake
this work.
- With this RFC, the workload (supply-chain) controller would have a
  dependency on the Runnable CRD.

# Alternatives
[alternatives]: #alternatives

An alternative is to drop the Runnable object altogether. Cartographer could
directly manage the creation and reporting of objects

# Prior Art
[prior-art]: #prior-art

Runnable and ClusterRunTemplate are the prior art.

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

## Resource reporting on workload
The workload reports the resources created from each step. Should this point
to the Runnable object or to the most recently successful immutable object
created?
- Pointing to a runnable may be surprising, as the user may not know what
  this is.
- Pointing to the most recently successful may be tricky. Is this
  information exposed to the user at all currently? Perhaps Runnable should
  be updated to report this information in its status (similar to workload
  reporting the resources).

## Simplify runnables that use 'selected'
ClusterRunTemplates are able to read values from one object that is
pre-existing in the cluster. They do this by specifying values to read from
'selected'. The selected object is the object that satisfies the Runnable's
.spec.selector field. Should we enable users that are leveraging this
behavior to use the simplification specified in this RFC?

## Success criteria
Runnable currently hard codes success to be a .status.condition with
Type:Succeeded and Status:True. E.g. we hard code Runnable to work with
Tekton. At some point we will likely want to allow users to specify a Runnable
success criteria for resources other than Tekton. Perhaps at that point we
will also have adopted success criteria for our other templates (as mooted
in the [Read Resources Only When In Success State RFC](https://github.com/vmware-tanzu/cartographer/pull/556))

## ClusterDeploymentTemplate
Is there need for this behavior on ClusterDeploymentTemplates?

## Warn on name collisions
The process above ensures that there will be no name collisions between
any set of auto-generated ClusterRunTemplates. But it cannot protect against
a name collision with a ClusterRunTemplate that has been manually created by
some app operator. Cartographer should be able to use labels to determine if
a ClusterRunTemplate has been auto-generated. Should Carto warn/refuse to
overwrite a ClusterRunTemplate if it determines that it has not been
autogenerated?

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

This RFC proposes adding an enum field named `lifecycle` to the .spec of:
- ClusterSourceTemplate
- ClusterImageTemplate
- ClusterConfigTemplate
- ClusterTemplate
