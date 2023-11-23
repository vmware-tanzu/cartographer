# Draft RFC <0000> move off monorepo

## Summary

Following the [convention proposal](https://github.com/vmware-tanzu/cartographer/pull/514) and [office hours discussion](https://www.youtube.com/watch?v=iCURQqV52Uw&t=1173s) there was a decision to not include conventions in the https://github.com/vmware-tanzu/cartographer repository as using the current cartographer repo as a monorepo would cause challenges around:

- extra support burden on the current maintainers
- maintaining development and style consistency
- existing CI, release automation and GitHub tooling doesn't work well with monorepos
- as we continue to build up the Cartographer ecosystem with integrations, adaptors, platform controllers this will cause a bottleneck for changes proposed on a single repo.

Adopting a microservice approach for Cartographer will aid extensions and help build up the wider ecosystem and community around it.
This RFC proposes how we will adopt this microservice style approach that will help grow the Cartographer ecosystem, encourage new integrations and extensions is a sustainable and maintainable way that suits the project and its aspirations.

## Motivation

The motivation behind Cartographer as an OSS project aims to build a community and ecosystem to aid supply chain choreography.  Cartographer itself is an "umbrella" for this ecosystem, the first step of this was to build the core components and now we are looking ahead to extend by adding more capabilities.
 
As expressed by maintainers of the https://github.com/vmware-tanzu/cartographer repository it will become hard to manage it if we use it as a monorepo and continue to add more capabilities like conventions etc.
It is expected that this RFC describes how we split the current repository so it allows the Cartographer OSS project to evolve and scale beyond the concrete capabilities the core aims to provide.
For example splitting out from the current repo:

-  the __`Runnable`__ feature so it can release independently and install standalone in a target deployment cluster,  avoids deploying all of Cartographer which contains build cluster CRDs not desired in runtime clusters
- the __examples directory__ so we can build up a wide range of examples that include integrations with wider platform services like conventions.  This also helps encourage contribution of community supply chains for different integrations where we can add automated testing that runs against different setups.
- the __website__ so we can add documentation for these new integrations and platform capabilities outside of the core repository.  I.e. Documenting services like the conventions or a Jenkins integration should not require a change to the core cartographer repository, this should live outside in its own repository.

## Possible Solutions

There seems to be only two options possible solutions:
- __monorepo__ we could revisit the [convention proposal](https://github.com/vmware-tanzu/cartographer/pull/514) decision and look at ways to address the monorepo concerns however there seemed a general consensus on the [office hours call](https://www.youtube.com/watch?v=iCURQqV52Uw&t=1173s) that this was not desired.
- __microservice__ As conventions are now being worked on in a different repository, we should look at splitting out parts that currently reside in https://github.com/vmware-tanzu/cartographer.

Suggestions:

-  __`Runnable`__ feature so it can install standalone in a target deployment cluster and avoids deploying all of Cartographer
-  __examples directory__ so we can build up a wide range of examples that include integrations with wider platform services like conventions.
-  __website__ so we can add documentation for these new integrations and platform capabilities outside of the core repository.  I.e. documenting a service like the convention components or a Jenkins integration should not require a change to the core cartographer repository, this should live outside in it's own repository
- __others?__

## Cross References and Prior Art

Carvel: https://github.com/vmware-tanzu - Corporate sponsored organisation
Kubernetes: https://github.com/kubernetes - Cloud Native Computing Foundation sponsored organisation
Tekton: https://github.com/tektoncd - Continuous Delivery Foundation sponsored organisation
 
# Unresolved Questions

[unresolved-questions]: #unresolved-questions

- Should we keep the rfc folder in the cartographer repo as it should relate to only the core components?
- Should we move the RFC process outside into its own repository too so that it provides governance across the wider project?
- Should we consider an enhancement process for the project and keep the current repo level RFC process?
Note examples of OSS projects that use an enhancement process:

- __KEP__ - https://github.com/kubernetes/enhancements/
- __TEP__ - https://github.com/tektoncd/community/tree/main/teps
- __JEP__ - https://github.com/jenkinsci/jep
