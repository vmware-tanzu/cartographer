# RFC 0014 Change Tracking

## Summary
Implement a change tracking feature to make it possible for clients to track input data as it traverses through the supply chain.
Motivation
Context
At the core of the architecture of cartographer is the concept that a supply chain is the orchestration of objects and their controllers.  

A supply chain, as the name suggests, chains together a set of objects defining how the (status) fields of one object feeds into the spec (and sometimes data) fields of another.  Thus creating an ordered chain of interacting objects.  Because controllers continuously reconcile their objects towards a desired state.  A supply chain is, therefore, able to choreograph an otherwise set of independent objects (and their controllers).

A workload acts as input into this supply chain seeding initial values into an object to make the first controller do work.  The supply chain manages this workload input data and outputs from each object as it propagates through the supply chain.

It is important to note however, that a workload can have multiple inputs and input into several objects in the supply chain all at the same time.  As a result when a workload changes, it may cause several controllers to do work simultaneously.  

It is also important to note that a controller may also do work outside of the work that the supply chain is choreographing.  Kpack images, for example, are often choreographed as part of a supply chain.  But the kpack controller may also do work in response to a base OS image update.  Nothing to do with the supply chain but impacting it none-the-less.  As a result the supply chain may be triggered part way through by this input.
Problem
Whilst, on the one hand, the behavior described above is virtuous and generally beneficial to automated outer loop workflows.  On the other hand, this behavior can be problematic for more user-centric, imperative workflows, such as those often found in the inner loop.  Those of debugging and live update.  

Providing a real world example.  Given a supply chain that has several inputs, let’s say source code image and a debug flag.  When a developer applies a workload after changing their source code and turning on debug.  Several services in the supply chain may trigger at the same time and these inputs will traverse through the supply chain at different times.  But the developer, initiating the debug session will want to wait for that specific source code change to arrive “in cluster” in an image prepped for debugging before attaching their debugger.  So, it is important to know when both inputs have fully traversed the supply chain.

## Possible Solutions
As the cartographer is already generally aware of workload inputs and outputs (artifacts) and how those artifacts from one component (service) are inputs into another.  It should, therefore, be possible to output, as additional status, the relationship between these artifacts and the service that produced it.

For example:

```
status:
   outputs:
   - type: image
     value: abc123
     passed: [image-builder]
     from:
     - type: source
       value: def567
       passed: [workload, source-cloner, unit-tester]
```

Here we see the run image `abc123`, produced by the `image-builder` service, and “containing” the source code image `def567`. 

Using the above example again when a workload is applied with the cli, updating both the source image and debug flags, it can then wait for the workload status to be Ready and for an output status showing the last artifact in the supply chain contains the original source image and the debug flag values we expect.

## Concerns/Questions
How much does a consumer of this output status need to understand about supply chains?  Does a consumer have to “know” what the last artifact produced by a supply chain is?  And how?  

## Cross References and Prior Art
kubectl `wait` suggests a general need to impose imperative outcomes on top of the eventually consistent k8s system.
