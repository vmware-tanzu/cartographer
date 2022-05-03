This document provides an overview of the major themes driving Cartographer development, as well as constituent features and capabilities. Most items are gathered from the community or include a feedback loop with the community. We hope this will serve as a reference point for Cartographer users and contributors to help us prioritize existing features, provide input on unmet needs, and understand where the project is heading, especially if a contribution may be conflicting with longer term plans. 

If it does conflict, the team and community will need to determine whether to adjust the roadmap or recommend changes to the contribution idea.

**How to Help**

Discussion on the roadmap can take place in threads under [Issues](https://github.com/vmware-tanzu/cartographer/issues) or in [community meetings](https://docs.google.com/document/d/1HwsjzxpsNI0l1sVAUia4A65lhrkfSF-_XfKoZUHI120/edit?usp=sharing). Please open and comment on an issue if you want to provide suggestions, use cases, and feedback to an item in the roadmap. Community members are encouraged to be actively involved, and also stay informed so contributions can be made with the most positive effect and limited duplication of effort.
 
If you'd like to contribute (thank you!!) but don't have anything to propose (all good!), check out our issues for issues tagged [good-first-issue](https://github.com/vmware-tanzu/cartographer/labels/good%20first%20issue) or say hi in the [#cartographer](https://kubernetes.slack.com/archives/C02HKPSEKV1) Slack channel - we’re always happy to have an impromptu chat!.

**How to add an item to the roadmap**

Please create an issue using the [Feature Request](https://github.com/vmware-tanzu/cartographer/issues/new?assignees=&labels=&template=feature_request_template.md) template to propose a feature for the project. We rely on our community to help us detail out and prioritize our efforts to improve Cartographer. 
Also feel free to submit a pull request against your issue for feedback from the  community!

**Current roadmap**

The following table includes the current roadmap for Cartographer. Please take the timelines & dates as proposals and goals. Priorities and requirements change based on community feedback, roadblocks encountered, community contributions, etc. If you depend on a specific item, we encourage you to attend community meetings to get updated status information, or help us deliver that feature by contributing!
 
April to June 2022



| Theme | Title | Outcome | Issue | RFC |
| ------| ----- | ------- | ----- | --- |
| Use Cases | Kubernetes resources as deployment targets | Provide an example deploying raw Kubernetes resources, not just knative apps | [#593](https://github.com/vmware-tanzu/cartographer/issues/593) | |
| Use Cases | Adding workload configuration for Maven artifacts | Users can use Maven artifacts as an input to their supply chains, not just images or source code | [#741](https://github.com/vmware-tanzu/cartographer/issues/741)| [#767](https://github.com/vmware-tanzu/cartographer/pull/767)|
|Visibility | Publish Kubernetes events from supply chain | Users can better understand and troubleshoot their supply chains (?) | [#711](https://github.com/vmware-tanzu/cartographer/issues/711) | [#756](https://github.com/vmware-tanzu/cartographer/pull/756) |
| Visibility | Allow resources to surface status and errors up to the workload | Users can better understand and troubleshoot their supply chains | [#694](https://github.com/vmware-tanzu/cartographer/issues/694) | [#738](https://github.com/vmware-tanzu/cartographer/pull/738)|
| Visibility | Artifact tracing | Users can know that an input to a resource produced a specific output | [#744](https://github.com/vmware-tanzu/cartographer/issues/744) | [#799](https://github.com/vmware-tanzu/cartographer/pull/799) |
| Adoption | Users can leverage an OOTB (Out-of-the-Box) supply chain in TCE (Tanzu Community Edition)| Users can get started with TCE without having to write their own supply chain from scratch | [#779](https://github.com/vmware-tanzu/cartographer/issues/779) |N/A |
|Adoption |Improve docs and site UX | Users have a great experience learning Cartographer - lots of great first-issues here! | [#17](https://github.com/vmware-tanzu/cartographer-site/issues/17), [#566](https://github.com/vmware-tanzu/cartographer/issues/566), [#610](https://github.com/vmware-tanzu/cartographer/issues/610), [#438](https://github.com/vmware-tanzu/cartographer/issues/438), [#544](https://github.com/vmware-tanzu/cartographer/issues/544) |N/A |
|Quality |Blueprint architecture / API evolution: Simplify architecture for our users | Reduce the number of Cartographer specific objects and YAMLs needed to set up a supply chain (Proposal only for this quarter) | [#731](https://github.com/vmware-tanzu/cartographer/issues/731) | [#766](https://github.com/vmware-tanzu/cartographer/pull/766) |
| Quality | Reduce duplication in core code | Make our implementation more elegant, lightweight, and secure | TBD |
| Quality | Define ways to improve the release process | |[#15](https://github.com/vmware-tanzu/cartographer/issues/15) |
| <Next> | <Your suggestion here!> |  | | |

 
**Other areas being explored**

* Connecting K8s objects to workloads (secrets, config maps, volumes, etc)
* Supply Chain supported Delivery verification (support smoke test on deployment)
* Alerting
* Fan In / Fan Out - from git repo to multiple clusters for different purposes, from multiple clusters to a single git repository. ex staging cluster + performance test cluster
* CloudEvent trigger - cleaner interoperability between different types of tooling, recieving CloudEvents ex if a Nexus supports webhooks, we dont have to do polling, like the image registry fires events
