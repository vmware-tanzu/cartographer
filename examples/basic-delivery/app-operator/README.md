# App Operator

In this directory you'll find Kubernetes objects necessary for the app
operators to submit to Kubernetes:

- a secret with credentials for a git repository
- a cluster source template that wraps a GitRepository object
- a cluster deployment template `app-deploy` which instantiates a `App` from configuration
  - The use of `App` here is important because of how `knative` updates the
    knative service under the hood to include some extra annotations that _can't_
    be mutated once applied by knative's controller. As `kapp` can be
    configured to not patch certain features (something `cartographer` can't
    yet), we're able to bridge that gap with the use of `kapp-ctrl/App`.
- a delivery that references these templates
- a deliverable that matches the delivery
