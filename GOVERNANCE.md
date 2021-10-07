
# Cartographer Governance

This document defines the project governance for Cartographer, an open source project by VMware.

## Overview

Cartographer is a Choreographer for Kubernetes. It allows users to configure K8s resources into re-usable [_Supply Chains_](site/content/docs/reference.md#ClusterSupplyChain) that can be used to define all of the stages that an [_Application Workload_](site/content/docs/reference.md#Workload) must go through to get to an environment. 

We are committed to the project not only delivering the distribution, but also building an open, inclusive, and productive vendor driven open source community; together, we will advance a reliable, nimble, and extensible foundation for modern applications.

## Community

**Users**: Members who engage with the Cartographer community, providing valuable feedback and unique perspectives.

**Contributors**: Members who contribute to the project through documentation, code reviews, responding to issues, participation in proposal discussions, contributing code, etc. The project welcomes code contributions to the Cartographer project, as well as contributor-originated packages that add capabilities from other projects. These contributed packages will conform to the Cartographer packaging requirements and lifecycle management.

**Maintainers**: The Cartographer project leaders are currently all employees of VMware. They are responsible for the overall health and direction of the project; final reviewers of PRs and responsible for releases. Some maintainers are responsible for one or more components within a project, acting as technical leads for that component. Maintainers are expected to contribute code and documentation, review PRs including ensuring quality of code, triage issues, proactively fix bugs, and perform maintenance tasks for these components. If a maintainer leaves VMware, they will also leave their maintainer position. As the project matures and more contributors are getting involved, this section might be revised.

## Request for Comment (RFC) Process

One of the most important aspects in any open source community is the concept of requests for comments (RFC). All large changes to the codebase and/or new features, including ones proposed by maintainers, should be preceded by an RFC in this repository. This process allows for all members of the community to weigh in on the concept (including the technical details), share their comments and ideas, and offer to help. It also ensures that members are not duplicating work or inadvertently stepping on toes by making large conflicting changes.

The project roadmap is defined by accepted RFCs.

RFCs should cover the high-level objectives, use cases, and technical recommendations on how to implement. In general, the community member(s) interested in implementing the RFC should be either deeply engaged in the RFC process or be an author of the RFC. Contributors are encouraged to refer to the [RFC proposal process](https://github.com/vmware-tanzu/cartographer/blob/main/rfc/README.md) for further details when creating new and reviewing existing RFCs. 

## Lazy Consensus

To maintain velocity in a project as busy as Cartographer, the concept of Lazy Consensus is practiced. Ideas and / or proposals should be shared by maintainers via GitHub with the appropriate maintainer groups (e.g., @vmware-tanzu/cartographer-maintainers) tagged. Out of respect for other contributors, major changes should also be accompanied by a ping on Slack, and a note on the Cartographer mailing list as appropriate. Author(s) of proposals, pull requests, issues, etc. will specify a time period of no less than five (5) working days for comment and remain cognizant of popular observed world holidays. Other maintainers may request additional time for review, but should avoid blocking progress and abstain from delaying progress unless absolutely needed. The expectation is that blocking progress is accompanied by a guarantee to review and respond to the relevant action(s) (proposals, PRs, issues, etc.) in short order. All pull requests need to be approved by two (2) maintainers.

Lazy Consensus is practiced for the main project repository and the additional repositories listed above.