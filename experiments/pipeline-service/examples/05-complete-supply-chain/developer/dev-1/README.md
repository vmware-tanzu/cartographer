# dev-1

The `dev-1` development team wants to continuously deliver their application
from two branches: 

- `master`
- `dev` 

but doesn't want to submit a [tekton/Pipeline] to match each
[carto.run/Workload] as, after all, it's the same codebase.

In order to do that, the team adds a shared label (that's unique to this team)
to both carto.run/Workloads (`pipelines.carto.run/group`) so that pipeline-service can
find that shared tekton/Pipeline when stamping out [tekton/PipelineRun].

See:

- `carto.run/Workload` for `master` branch:
  [workload-master.yaml](./workload-master.yaml)
- `carto.run/Workload` for `dev` branch:
  [workload-dev.yaml](./workload-dev.yaml)
- `carto.run/ClusterSourceTemplate` where `carto.run/Pipeline` selects based on
  those labels:
  [supply-chain-templates.yaml](../../app-operator/supply-chain-templates.yaml)

[tekton/Pipeline]: https://github.com/tektoncd/pipeline/blob/51f3ce8f36e605724bad3057d9a9b621bdb4df8e/docs/pipelines.md#configuring-a-pipeline
[tekton/PipelineRun]: https://github.com/tektoncd/pipeline/blob/51f3ce8f36e605724bad3057d9a9b621bdb4df8e/docs/pipelineruns.md
[carto.run/Workload]: http://carto.run/docs/reference/#workload
