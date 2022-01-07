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
  '── shared | testing-sc
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
  # the git server, for example gitlab.com
  server: github.com
  # the project and repo name
  repository: example/example.git
  # private ssh key. Must match a public key on the git server
  base64_encoded_ssh_key: a-key
  # public key of the git server. Your ~/.ssh/known_hosts may have examples
  base64_encoded_known_hosts: a-host
  # the branch to which configuration will be pushed
  branch: main
  # username in the git server
  username: example
  # user email
  user_email: example@example.com
  # port of the git server
  port: ""
  # ssh name for accessing the git server
  ssh_user: git
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
    apiVersion: v1
    kind: ConfigMap # <=== a configmap is templated
    metadata:
      name: #@ data.values.workload.metadata.name # <=== The same templating context is available
                                                         with info about the workload, params and artifacts
                                                         created earlier by the supply chain
    data:
      #@yaml/text-templated-strings
      manifest: |
        apiVersion: serving.knative.dev/v1
        kind: Service # <=== in the configmap.data field, a Service object is configured
        metadata:
          name: (@= data.values.workload.metadata.name @)
          labels:
            (@- if hasattr(data.values.workload.metadata, "labels"): @) # <=== an optional label for kapp controller
                                                                               is written if specified by the app dev
            (@- if hasattr(data.values.workload.metadata.labels, "app.kubernetes.io/part-of"): @)
            app.kubernetes.io/part-of: (@= data.values.workload.metadata.labels["app.kubernetes.io/part-of"] @)
            (@ end -@)
            (@ end -@)
            carto.run/workload-name: (@= data.values.workload.metadata.name @)
            app.kubernetes.io/component: run
        spec:
          template:
            metadata:
              annotations:
                autoscaling.knative.dev/minScale: "1"
            spec:
              containers:
                - name: workload
                  image: (@= data.values.image @)
                  securityContext:
                    runAsUser: 1000
              imagePullSecrets:
                - name: registry-credentials
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
    - name: git_writer_username
      default: example            # <=== specified as `default` this value can be overwritten by the workload
                                  #      if specified as `value`, it cannot be overwritten
    - name: git_writer_user_email
      default: example@example.com
    - name: git_writer_commit_message
      default: "Update app configuration"
    - name: git_writer_ssh_user
      default: git
    - name: git_writer_server
      default: github.com
    - name: git_writer_port
      default: ""
    - name: git_writer_repository
      default: example/example.git
    - name: git_writer_branch
      default: main
    - name: git_writer_skip_host_checking
      default: false
    - name: git_writer_ssh_variant
      default: ssh
  template:
    apiVersion: carto.run/v1alpha1
    kind: Runnable # <=== the cluster template creates a runnable
    metadata:
      name: $(workload.metadata.name)$-git-writer
    spec:
      runTemplateRef:
        name: git-writer # <=== the cluster run template that specifies the object to be created

      inputs: # <=== these inputs are passed to the cluster run template
        input_config_map_name: $(workload.metadata.name)$
        input_config_map_field: manifest.yaml # <=== the filename that will be read by the delivery specified here

        git_username: $(params.git_writer_username)$
        git_user_email: $(params.git_writer_user_email)$
        commit_message: $(params.git_writer_commit_message)$
        git_ssh_user: $(params.git_writer_ssh_user)$
        git_server: $(params.git_writer_server)$
        git_server_port: $(params.git_writer_port)$
        git_repository: $(params.git_writer_repository)$
        branch: $(params.git_writer_branch)$
        skip_host_checking: $(params.git_writer_skip_host_checking)$
        git_ssh_variant: $(params.git_writer_ssh_variant)$
        data: $(config)$ # <=== here is the data, the config yaml
```

The cluster run template creates a tekton taskrun. All tekton taskruns
provide values to a tekton task. For simplicity sake, the example uses
the [git-cli task defined in the tekton catalog](https://github.com/tektoncd/catalog/tree/main/task/git-cli/0.2).

As documented in the tekton catalog, this task expects a number of parameters
and workspaces to be defined in the taskrun.

One workspace is a secret object with credentials for the git operations:

```yaml
#@ load("@ytt:data", "data")

---
apiVersion: v1
kind: Secret
metadata:
  name: git-ssh-secret
data:
  id_rsa: abc123 # <=== see https://github.com/tektoncd/catalog/tree/main/task/git-cli/0.2#using-ssh-credentials
  known_hosts: xyz456
```

Here we see the cluster run template which creates taskruns that fulfill
the `git-cli` task contract.

```yaml
apiVersion: carto.run/v1alpha1
kind: ClusterRunTemplate
metadata:
  name: git-writer
spec:
  template:
    apiVersion: tekton.dev/v1beta1
    kind: TaskRun
    metadata:
      generateName: $(runnable.metadata.name)$-
    spec:
      taskRef:
        name: git-cli
      workspaces:
        - name: source
          emptyDir: { }
        - name: input
          emptyDir: { }
        - name: ssh-directory
          secret:
            secretName: git-ssh-secret # <=== the secret created above
      params:
        - name: GIT_USER_NAME
          value: $(runnable.spec.inputs.git_username)$
        - name: GIT_USER_EMAIL
          value: $(runnable.spec.inputs.git_user_email)$
        - name: USER_HOME
          value: /root
        - name: GIT_SCRIPT # <=== the git script below creates a commit in the repo
                           #      with the contents of the earlier configmap
          value: |
            export COMMIT_MESSAGE="$(runnable.spec.inputs.commit_message)$"
            export BRANCH="$(runnable.spec.inputs.branch)$"
            if [[ -n "$(runnable.spec.inputs.skip_host_checking)$" && "$(runnable.spec.inputs.skip_host_checking)$" = true ]]
            then
              export GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
            fi
            if [[ -n "$(runnable.spec.inputs.git_ssh_variant)$" ]]
            then
              export GIT_SSH_VARIANT="$(runnable.spec.inputs.git_ssh_variant)$"
            fi
            git init
            if [[ -n "$(runnable.spec.inputs.git_server_port)$" ]]; then
              git remote add origin $(runnable.spec.inputs.git_ssh_user)$@$(runnable.spec.inputs.git_server)$:$(runnable.spec.inputs.git_server_port)$/$(runnable.spec.inputs.git_repository)$
            else
              git remote add origin $(runnable.spec.inputs.git_ssh_user)$@$(runnable.spec.inputs.git_server)$:$(runnable.spec.inputs.git_repository)$
            fi
            # TODO remove the fetch and branch
            git fetch
            git branch
            git pull origin "`git remote show origin | grep "HEAD branch" | sed 's/.*: //'`"
            git pull origin "$BRANCH" || git branch "$BRANCH"
            git checkout "$BRANCH"
            export CONFIG_MAP_FIELD=$(runnable.spec.inputs.input_config_map_field)$
            export DATA="$(runnable.spec.inputs.data)$" # <===
                                                        #      the data (the config yaml) is written to the
                                                        #      specified file
            echo "$DATA" | tee "$CONFIG_MAP_FIELD"      # <===
            git add .
            git commit --allow-empty -m "$COMMIT_MESSAGE"
            git push --set-upstream origin "$BRANCH"
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
