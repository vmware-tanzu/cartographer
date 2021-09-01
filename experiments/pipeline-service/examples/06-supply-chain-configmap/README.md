# test configuration via configmaps

In this example we illustrate how operators could've configured the mechanism
for developers to supply their testing configuration via [ConfigMap]s rather
then [tekton/Pipeline], considerably lowering the bar.


```
SUPPLYCHAIN

   ...............              ............... 
   source-provider <----src---- pipeline-runner 
   ...............              ............... 
         .                            .
         .                            .
         .                            .
 fluxcd/GitRepository        carto.run/Pipeline
                                      |
                                      '
                              tektoncd/PipelineRun[1...n]

```


[ConfigMap]: https://kubernetes.io/docs/concepts/configuration/configmap/
[tekton/Pipeline]: https://github.com/tektoncd/pipeline/blob/main/docs/pipelines.md

