# Draft RFC 0016 Validate Template Output

## Summary

Template authors specify fields that should be exposed from the template to
later templates in the Supply Chain graph. e.g. SourceTemplates expose a
url and a revision, by specifying a path where these values can be read.
Currently, no validation is done to ensure that the values are sensible.
Cartographer should undertake validation and should report a bad condition
when reading bad values.

## Motivation

### Throw the right errors

Errors should be thrown as close to the cause as possible, and should
identify the cause.
> Each exception that you throw should provide enough context to determine
> the source and location of an error. Create informative error messages and
> pass them along with your exceptions. Mention the operation that failed and
> the type of failure.

_Robert C. Martin: Clean Code_

At present, a template can expose a bad value. That problem will not be
revealed until a later in template in the supply chain graph attempts to use
the value to stamp out an object. The stamped object would then be rejected by
the apiServer. Cartographer would specify a creation error. Users would have to
hunt down the true root cause, a SourceTemplate exposing bad values. 

### Cartographer is not a continuous thing-doer

Without validation, Cartographer can be used to pass any values between any
objects. Without validation, template types are little more than syntactic sugar
for generic templates. e.g.

ClusterSourceTemplate => a template that exposes 2 values
ClusterImageTemplate, ClusterConfigTemplate => templates that expose 1 value
ClusterTemplate => a template that exposes 0 values

By adding validation to outputs, ImageTemplates and ConfigTemplates would become
truly typed objects with different behavior.

## Proposed Changes

### Per template validation

The template interface has a GetOutput method. This is how values are collected
from one template to pass to another. The interface should similarly have a
ValidateOutput method. Each template type should implement its own validation. e.g.
ConfigTemplates would validate that their output is valid yaml for a valid k8s object.

#### Examples
Below are concrete examples of templates that would pass and fail validation.
These examples are not meant to be exhaustive.

##### ClusterSourceTemplate
The url and revision fields to which a `ClusterSourceTemplate` points should resolve to
strings. Further validation (ensuring that the string is a valid uri) is appropriate.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: fails-validation
spec:
  urlPath: .data.some_malformed_uri
  revisionPath: .data.some_malformed_revision
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_malformed_uri: 6
      some_malformed_revision:
        key: this-is-an-object-not-a-string

---
apiVersion: carto.run/v1alpha1
kind: ClusterSourceTemplate
metadata:
  name: passes-validation
spec:
  urlPath: .data.some_uri
  revisionPath: .data.some_revision
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_uri: https://github.com/ossu/computer-science.git
      some_revision: main
```

##### ClusterImageTemplate
The image field to which a `ClusterImageTemplate` points should resolve to a
string. Further validation (ensuring that the string is a valid digest) is appropriate.

```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: fails-validation
spec:
  imagePath: .data.some_malformed_image_digest
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_malformed_image_digest: 6

---
apiVersion: carto.run/v1alpha1
kind: ClusterImageTemplate
metadata:
  name: passes-validation
spec:
  imagePath: .data.some_image_digest
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_image_digest: index.docker.io/your-project/app@sha256:db6697a61d5679b7ca69dbde3dad6be0d17064d5b6b0e9f7be8d456ebb337209
```

##### ClusterConfigTemplate
The config field to which a `ClusterConfigTemplate` points should resolve to a
valid k8s object.

  ```yaml
---
apiVersion: carto.run/v1alpha1
kind: ClusterConfigTemplate
metadata:
  name: fails-validation
spec:
  configPath: .data.some_invalid_object_definition
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_invalid_object_definition:
        kind: some-kind
        other-field: not-enough-info-for-object-specification

---
apiVersion: carto.run/v1alpha1
kind: ClusterConfigTemplate
metadata:
  name: passes-validation
spec:
  configPath: .data.some_k8s_object
  template:
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: test-configmap
    data:
      some_k8s_object:
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: turtles-all-the-way-down
        data:
          turtles: turtles
```

## Alternatives

### No validation, but greater user visibility

[RFC 3](rfc-0003-intermediate-value-crds.md) and [RFC 14](rfc-0014-change-tracking.md)
both discuss the need to give Cartographer users and dependencies greater visibility
in the values being passed between objects. Cartographer could substitute this greater
user visibility for better validation. This was rejected as insufficient.

## Implementation

The [template interface](pkg/templates/templates.go) includes the method `GetOutput() (*Output, error)`
Each template implements this method. In these respective implementations, before returning
an output, the template should validate that the value to be returned passes validation.
Otherwise, a helpful error should be returned.
