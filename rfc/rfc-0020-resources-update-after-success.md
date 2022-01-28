# Draft RFC 20 Update Resources Only After Success/Failure

## Summary

Cartographer is currently unable to determine when an input to a resource has reached success, failed, or is still
processing. Making Carto aware of status indications is not sufficient, as frequent updates to a resource could keep it
in a constant state of 'processing', even as it succeeds and fails on multiple successive inputs. Carto should wait
until an input has resulted in success or failure before updating the resource with new input. 

## Motivation

Connecting an output of a resource to an input is necessary for establishing provenance. That is to state, "The app 
currently running on the cluster is a result of commit X," it is necessary to tie a resource output to the input that
produced it. Waiting for success/failure before update achieves this.

Tieing inputs to outputs also allows Carto to harden supply chains to tampering. Currently, an enemy could update a
resource. Carto will not read said resource without first updating it with the proper spec. But the resource may still
produce the enemy output (followed by the correct output). Carto must be able to associate inputs to outputs if it is
not to propagate the enemy output.

## Proposed Solution

### Use ObservedCompletion and ObservedMatch (from DeploymentTemplate)

Templates that expose outputs (SourceTemplate, ImageTemplate...) can include success conditions. These can include
expected values at a given path on the stamped object. These can alternately be an expectation that one field (e.g.
on the status) matches another field (e.g. on the spec).

- An observed completion: includes a mandatory SucceededCondition and an optional FailedCondition. Both conditions are 
  defined by a path and value. When the object's observedGeneration == generation, and the value at the
  specified path matches the stated value, then this condition is met. 
- An observed matches: a list of matches. Each match is a definition of two paths. When the values at the two paths are
  the same, then this condition is met.

### Example

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterBuildTemplate
metadata:
  name: example-build---consume-output-of-components
spec:
  template:
    apiVersion: kpack.io/v1alpha1
    kind: Image
    metadata:
      name: ...
    spec:
      ...
  imagePath: $(status.latestImage)$
  observedCompletion:
    succeeded:
      - key: status.conditions[?(@.type=="Ready")].status
        value: True
```

### Handling non-reconciling objects

Some objects are not reconciled and are immediately valid (for example, configmaps). How should these be handled?

1. Users could write trivial ObservedMatches. For example asserting that the `.data` field is equal to the `.data`
   field.
2. The absence of either an ObservedCompletion or an ObservedMatch could be taken as indication that the object
   being created is immediately succesful. (This seems dangerous and ill-advised)
3. An additional exclusive field (one that could be specified instead of ObservedMatches and ObservedCompletion) could
   be defined. This field could be `AlwaysSuccessful: True`

### Limitations

There are a few limitations to the current setup of observedCompletion and ObservedMatches:
1. ObservedCompletion is limited to matching a single path and value. If more than one path must be interrogated,
   this spec is not sufficient.
2. ObservedMatches cannot define a failure state.

## Possible Extensions

### Allow boolean operations

- An OR condition: An OR contains a list of conditions. When any is met, the OR condition is met.
- An AND condition: A list of conditions. When all conditions are met, this condition is met. Meant primarily for
  nesting in OR conditions (as AND is the default relation of a list of conditions).
- A NOT condition: holds a condition. The NOT condition is true only when its condition is false.

### Read other objects on the cluster

It may be useful to compare the stamped object to another object on the cluster. Or to simply read a value from
another object on the cluster. 

## Cross References and Prior Art

The Deployment Template currently requires either an `ObservedMatches` or `ObservedCompletion` field.
