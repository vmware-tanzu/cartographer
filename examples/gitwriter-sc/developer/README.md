# developer

Here you'll find all that the developer needs to submit to Kubernetes to have
its code going through the software supply chain defined by the app operators.

- a workload with
    - location of the source code
    - name of the service account that provides permission to deploy templated objects
    - environment variables for the build process
    - a label referenced in the service creation
- a service account with permission to create configmaps and tekton taskruns

Note: The examples/shared directory also contains developer resources.
