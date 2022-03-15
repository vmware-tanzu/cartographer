# Source to App Config In Git

**before you proceed**: the example in this directory illustrates the use of
the latest components and functionality of Cartographer (including some that
may not have been included in the latest release yet). Make sure to check out
the version of this document in a tag that matches the latest version (for
instance, https://github.com/vmware-tanzu/cartographer/tree/v0.0.7/examples).

---

The [basic-sc] example illustrates how an App Operator group could set up a software
supply chain such that source code gets continuously built using the best
practices from [buildpacks] via [kpack/Image] and deployed to the cluster using
[knative-serving]. This example will alter that final step; rather than deploy in the
cluster, configuration for deploy will be written to git. The [delivery] example
then picks up the configuration to deploy the app in another cluster.

```

  source --> image --> configuration --> git

```

As with [basic-sc], the directories here are structured in a way to reflect which Kubernetes
objects would be set by the different personas in the system:


```
  '── shared | gitwriter-sc
      ├── 00-cluster                         preconfigured cluster-wide objects
      │                                        to configured systems other than
      │                                              cartographer (like, kpack)
      │
      │
      ├── app-operator                      cartographer-specific configuration
      │   ├── supply-chain-templates.yaml            that an app-operator would
      │   ├── ...                                                        submit
      │   └── supply-chain.yaml
      │
      │
      └── developer                         cartographer-specific configuration
          ├── ...                                       that an app-dev submits
          └── workload.yaml
```

## Prerequisites

1. Install the prerequisites of the [basic-sc] example.

2. Install [tekton]

```bash
TEKTON_VERSION=0.30.0 kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v$TEKTON_VERSION/release.yaml
```

3. Install the [git-cli task](https://github.com/tektoncd/catalog/tree/main/task/git-cli/0.2) from the
  [tekton catalog](https://github.com/tektoncd/catalog). This is used to write to the git repo.

```bash
kapp deploy --yes -a tekton-git-cli -f https://raw.githubusercontent.com/tektoncd/catalog/main/task/git-cli/0.2/git-cli.yaml
```

## Running the example in this directory

### Location of files

As with [basic-sc], this example uses two directories with sub-directories of kubernetes resources:
[../shared](../shared) and [.](.).

The shared directory has a subdirectory for cluster-wide configuration: [../shared/cluster](../shared/cluster)

There is a subdirectory of cartographer-specific files that an App Operator would submit
in both this and in the shared directory:
[./app-operator](./app-operator) and [../shared/app-operator](../shared/app-operator)

Finally, in this and in the shared directory there is a subdirectory containing
Kubernetes objects that a developer would submit: [./developer](./developer) and
[../shared/developer](../shared/developer)

### Configuring the example

First, follow the [example configuration steps in basic-sc](../basic-sc/README.md#configuring-the-example)

Next, update [values.yaml](./values.yaml) with information about your git
setup.

```yaml
#@data/values
---
# configuration necessary for pushing the config to a git repository.
#
git_writer:
  # the git commit message when the config definition is committed to the repo
  message: "Update app configuration"
  # the git server, project, and repo name
  repository: github.com/example/example.git
  # the branch to which configuration will be pushed
  branch: main
  # username in the git server
  username: example
  # user email
  user_email: example@example.com
```

### Deploying the files

Similar to the [deploy instructions in basic-sc](../basic-sc/README.md#deploying-the-files):

```bash
kapp deploy --yes -a example -f <(ytt --ignore-unknown-comments -f .) -f <(ytt --ignore-unknown-comments -f ../shared/ -f ./values.yaml)
```

### Observing the example

When using tree, we see two additional resources, the configmap and the Runnable. As seen
in [runnable-tekton], the Runnable will have new taskrun children as the supply
chain propagates updates.

```console
$ kubectl tree workload dev
NAMESPACE  NAME
default    Workload/gitwriter-sc
default    ├─GitRepository/gitwriter-sc                  source fetching
default    ├─Image/gitwriter-sc                          image building
default    │ ├─Build/gitwriter-sc-build-1
default    │ │ └─Pod/gitwriter-sc-build-1-build-pod
default    │ ├─PersistentVolumeClaim/gitwriter-sc-cache
default    │ └─SourceResolver/gitwriter-sc-source
default    ├─ConfigMap/gitwriter-sc                      app configuration writing
default    └─Runnable/gitwriter-sc-git-writer            configuration commit to git
default      └─TaskRun/gitwriter-sc-git-writer-pqgjl
default        └─Pod/gitwriter-sc-git-writer-pqgjl-pod
```

After the runnable completes, the user can go to the git repository configured
and find a `manifest.yaml` file written in the chosen branch.

## Tearing down the example and uninstalling the dependencies

1. Delete tekton: ```TEKTON_VERSION=0.30.0 kubectl delete -f https://storage.googleapis.com/tekton-releases/pipeline/previous/v$TEKTON_VERSION/release.yaml```
2. Delete the tekton catalog git-cli: ```kapp delete -a tekton-git-cli```
4. Follow the [basic-sc teardown instructions](../basic-sc/README.md#tearing-down-the-example).
5. Manually delete any commits in the git repository that are no longer desired.

## Step by step

This example creates a software supply chain like mentioned above

```

  source --> image --> configuration --> git

```

In [basic-sc] the `image` is deployed directly in the cluster. Here, we will
create the configuration for the app and write the configuration to a git repository.

### Need: separating build and production environment

App operators may wish to limit the exposure of the build cluster to the outside
world. In this case, the apps created in the supply chain will be deployed in a
separate production cluster serving clients.

This supply chain takes app operator configuration of the git repository that
will be used to write the configuration. The location and even existence of this
repository can be abstracted away from the developer. After building an image, the
app configuration is written to a configmap.

### Creating the configuration

In this example, ytt templating is used in order to achieve conditional templating.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterConfigTemplate
metadata:
  name: app-config
spec:
  configPath: .data.manifest # <=== the yaml written below is exposed to the supply chain as `config`

  # The ytt field is a single formatted string.
  # This is necessary because ytt uses comments to template,
  # but objects submitted to the cluster will have comments removed.
  # By including the comments in the text field, this stripping of
  # context is avoided.
  ytt: |
    #@ load("@ytt:data", "data")
    #@ load("@ytt:json", "json")
    #@ load("@ytt:base64", "base64")

    #@ def manifest():
    manifest.yaml: #@ manifest_contents()
    #@ end

    #@ def manifest_contents():
    apiVersion: serving.knative.dev/v1
    kind: Service
    metadata:
      name: #@ data.values.workload.metadata.name
      labels:
        #@ if hasattr(data.values.workload.metadata, "labels"):  # <=== an optional label for kapp controller
                                                                        is written if specified by the app dev
        #@ if hasattr(data.values.workload.metadata.labels, "app.kubernetes.io/part-of"):
        app.kubernetes.io/part-of: #@ data.values.workload.metadata.labels["app.kubernetes.io/part-of"]
        #@ end
        #@ end
        carto.run/workload-name: #@ data.values.workload.metadata.name
        app.kubernetes.io/component: run
    spec:
      template:
        metadata:
          annotations:
            autoscaling.knative.dev/minScale: '1'
        spec:
          containers:
            - name: workload
              image: #@ data.values.image
              securityContext:
                runAsUser: 1000
          imagePullSecrets:
            - name: registry-credentials
    #@ end

    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: #@ data.values.workload.metadata.name # <=== The same templating context is available
                                                         with info about the workload, params and artifacts
                                                         created earlier by the supply chain
    data:
      manifest: #@ base64.encode(json.encode(manifest()))
```

### Writing the configuration to git

A cluster template is used to create a runnable. This runnable will create taskruns
to write each new configuration to git.

```yaml
#@ load("@ytt:data", "data")

apiVersion: carto.run/v1alpha1
kind: ClusterTemplate # <=== the cluster template exposes no values to the supply chain
metadata:
  name: git-writer
spec:
  params:
    - name: git_repository
      default: github.com/example/example.git    # <=== specified as `default` this value can be overwritten by the workload
                                                 #      if specified as `value`, it cannot be overwritten
    - name: git_branch
      default: main
    - name: git_user_name
      default: example            
    - name: git_user_email
      default: example@example.com
    - name: git_commit_message
      default: "Update app configuration"
  template:
    apiVersion: carto.run/v1alpha1
    kind: Runnable # <=== the cluster template creates a runnable
    metadata:
      name: $(workload.metadata.name)$-git-writer
    spec:
      serviceAccountName: default
      runTemplateRef:
        name: tekton-taskrun # <=== the cluster run template that specifies the object to be created

      inputs: # <=== these inputs are passed to the cluster run template
        serviceAccount: default
        taskRef:
          kind: ClusterTask
          name: git-writer
        params:
          - name: git_repository
            value: $(params.git_repository)$
          - name: git_branch
            value: $(params.git_branch)$
          - name: git_user_name
            value: $(params.git_user_name)$
          - name: git_user_email
            value: $(params.git_user_email)$
          - name: git_commit_message
            value: $(params.git_commit_message)$
          - name: git_files
            value: $(config)$  # <=== here is the data, the config yaml
```

The cluster run template creates a tekton taskrun. All tekton taskruns
provide values to a tekton task. Here we see the tekton cluster task that 
the taskrun is using.

```yaml
apiVersion: tekton.dev/v1beta1
kind: ClusterTask
metadata:
  name: git-writer
spec:
  description: |-
    A task that writes a given set of files (provided as a json base64-encoded)
    to git repository under a specific directory (`./config`).
  params:
    - name: git_repository
      description: The repository path
      type: string
    - name: git_branch
      description: The git branch to read and write
      type: string
      default: "main"
    - name: git_user_email
      description: User email address
      type: string
      default: "example@example.com"
    - name: git_user_name
      description: User name
      type: string
      default: "Example"
    - name: git_commit_message
      description: Message for the git commit
      type: string
      default: "New Commit"
    - name: git_files
      type: string
      description: >
        Base64-encoded json map of files to write to registry, for example -
        eyAiUkVBRE1FLm1kIjogIiMgUmVhZG1lIiB9
  steps:
    - name: git-clone-and-push
      image: paketobuildpacks/build:base
      securityContext:
        runAsUser: 0
      workingDir: /root
      script: |
        #!/usr/bin/env bash

        set -o errexit
        set -o xtrace

        git clone $(params.git_repository) ./repo
        cd repo

        git checkout -b $(params.git_branch) || git checkout $(params.git_branch)
        git pull --rebase origin $(params.git_branch) || true

        git config user.email $(params.git_user_email)
        git config user.name $(params.git_user_name)

        mkdir -p config && rm -rf config/*
        cd config

        echo '$(params.git_files)' | base64 --decode > files.json
        eval "$(cat files.json | jq -r 'to_entries | .[] | @sh "mkdir -p $(dirname \(.key)) && echo \(.value | tojson) > \(.key) && git add \(.key)"')"

        git commit -m "$(params.git_commit_message)"
        git push origin $(params.git_branch)
```

Here we see the cluster run template which creates taskruns that fulfill
the `git-writer` task contract.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: tekton-taskrun
spec:
  template:
    apiVersion: tekton.dev/v1beta1
    kind: TaskRun
    metadata:
      generateName: $(runnable.metadata.name)$-
      labels: $(runnable.metadata.labels)$
    spec:
      serviceAccountName: $(runnable.spec.inputs.serviceAccount)$
      taskRef: $(runnable.spec.inputs.taskRef)$
      params: $(runnable.spec.inputs.params)$
```

### Including the templates in the supply chain

Next we update the supply chain. We create a new step in the supply-chain, `config-provider` that
references the new ClusterConfigTemplate. The supply chain passes this resource the values exposed
by the `image-builder` step. The supply chain reads the `config` value exposed by this step and
passes it along to the new `git-writer` step. [See here](./app-operator/supply-chain.yaml).

In this way, the supply chain ensures that every update to the source code results in a new app
configuration written to a git repository.

[buildpacks]: https://buildpacks.io/
[knative-serving]: https://knative.dev/docs/serving/
[kpack/Image]: https://github.com/pivotal/kpack/blob/main/docs/image.md
[tekton]: https://github.com/tektoncd/pipeline
[basic-sc]: ../basic-sc/README.md
[runnable-tekton]: ../runnable-tekton/README.md
[delivery]: ../basic-delivery/README.md
