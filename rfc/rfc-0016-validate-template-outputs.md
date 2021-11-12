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

## Possible Solutions

### Per template validation

The template interface has a GetOutput method. This is how values are collected
from one template to pass to another. The interface should similarly have a
ValidateOutput method. Each template type should implement its own validation. e.g.
ConfigTemplates would validate that their output is valid yaml for a valid k8s object.

### No validation, but greater user visibility

[RFC 3](rfc-0003-intermediate-value-crds.md) and [RFC 14](rfc-0014-change-tracking.md)
both discuss the need to give Cartographer users and dependencies greater visibility
in the values being passed between objects. Cartographer could substitute this greater
user visibility for better validation.
