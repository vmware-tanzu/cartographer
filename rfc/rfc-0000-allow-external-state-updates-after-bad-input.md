# Meta
[meta]: #meta
- Name: Allow external state updates after bad inputs
- Start Date: 2022-02-03
- Author(s): Waciuma Wanjohi
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

Cartographer relies on resources to update objects with external state.
For example, kpack updates the most recent image when new base images are
available. But if Cartographer gives a bad input to an object, these updates
will no longer be applied. When a bad input to an object causes a failure,
Cartographer should create a shadow 'last good' object which can receive
external state updates until Cartographer passes inputs that cause the primary
object to re-enter a good state.

# Motivation
[motivation]: #motivation

Let us consider a supply chain that uses kpack to build images. After a
number of good commits, some bad commit is made. The kpack `Image` fails to
build the image. The `Image` continues to expose the `latestImage` field,
which is the most recent image created.

Then, a CVE in the base image is identified. kpack creates a new base image.
But the image will attempt to build an image with the current (bad) inputs. This
fails. The `Image` continues to expose the `latestImage` field,
which is the most recent image created. _This image still contains the CVE._

# What it is
[what-it-is]: #what-it-is

## Terminology

#### Failure
Failure indicates that an object is no longer attempting to reconcile and is in some bad state,
without useful outputs from the most recent set of inputs.

#### Success
Success indicates that an object has completed reconciling, is healthy and has useful outputs
from the most recent set of inputs.

### Shadow object
Assume that a supply chain has stamped out object A with input set N. Object A has reconciled
and is in a failed state. A shadow object is an object created from the same
workload/supply-chain/template but with the input set N-1 (or N-2, N-3...). This object would
need some name suffix in order to live in the cluster alongside the primary object.

### Primary object
The original object stamped from the most current set of inputs available to the template.

## Responsibilities

### Template Author
Define fields/values that indicate the object to be stamped is in a success or failure condition.

## Example

When an object fails, Cartographer creates a shadow object with the most recent good inputs.
External state updates will result in this object passing forward updated outputs. Carto uses
this shadow object until the primary object returns to a successful state.

# How it Works
[how-it-works]: #how-it-works

Given a template with defined failure state, assume that input set N is used to stamp the object.
The object then fails. Cartographer creates an additional shadow-object from the most recently
successful set of inputs (e.g. input set N-1). Cartographer passes the outputs of this object to the
next step in the supply chain. When the object's controller detects external state (e.g. when kpack
observes a new base image) it will reconcile this shadow object. The latest good output for the
step can therefore remain up to date with external changes (like CVE updates), even if some
inputs to the supply chain (e.g. a bad commit) cause the step to fail.

Eventually, some new input set (e.g. input set N+1; a new commit) will result in the object
entering a success state. Then the shadow object is destroyed. The outputs of the original object
are passed to the next step in the supply chain.

[A diagram of the current behavior and the proposed behavior](https://miro.com/app/board/uXjVOQeAuiU=/?invite_link_id=832395482222)

# Migration
[migration]: #migration

There should be no issue with a forward migration.

# Drawbacks
[drawbacks]: #drawbacks

This RFC complicates the design and behavior of Cartographer to handle a corner case. It increases the number
of objects in the cluster. When an object fails, it creates a second object in the cluster owned by the
same workload, from the same template, at the same step in the supply chain. This will likely be
confusing for users observing the tree of objects created by Cartographer.

How likely is this corner case? The example given is a bad commit that results in kpack being unable to
build an image. Would this be a problem that we would see on the path to prod? Would perhaps tests catch
the vast majority of such problems? Would we expect platforms built on top of Cartographer to have
development inner loops (which build images) and outer loops (which similarly build images) that would
result in the outer loop being protected from this concern?

# Alternatives
[alternatives]: #alternatives

## Put responsibility on the choreographed controllers

Cartographer can choose to wash its hands of this issue. Exposing a
`latestGood` field is a choice of the choreographed controller (e.g. kpack).
Is it the responsibility of the controller to ensure this image is still good?

## Put the responsibility on the user

The scenario in question occurs when the supply chain provides bad inputs
to a step in the chain. The example used is a code commit that cannot be built.
Should Cartographer worry that a supply chain will have bad inputs for so long
that CVEs in base images are found, patched, excluded in the supply chain output
and vulnerabilities persist in the supply chain output? Is it user responsibility
to notice when a supply chain step has failed and to address the problem?

## Use a rollback strategy, rather than a shadow object

Cartographer can keep the set of inputs that led to the last good input. When an
object reconciles and enters a failed state, Cartographer can rollback the definition
of the object to the last good inputs. New inputs would then be written on this object.

We would prefer the shadow object strategy if we expected that failed states might be
fleeting. That is if an object can get good inputs, report failure, then independently
kick off a re-reconcile and enter a good state, then we would want to enable that.
(Using a rollback strategy precludes such events.)

# Prior Art
[prior-art]: #prior-art

_Discuss prior art, both the good and bad._

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

The implementation of this feature could have implications for the implementation
of tracing. The specification of success and failure conditions must be agreed upon in order to
implement this RFC.

## Disambiguate bad inputs from bad external state?

The diagram of this proposal notes that there is little advantage to creating a shadow object if
an object's failure is due to the external state rather than due to bad inputs. Should the implementation
disambiguate these causes of object failure and only create shadow objects in the case of bad inputs?

# Spec. Changes (OPTIONAL)
[spec-changes]: #spec-changes

No changes to the spec is expected. (Success and failure conditions are assumed from other RFCs)
This should all be internal to Cartographer and observable only if a user is watching the objects
created and managed by a supply chain.
