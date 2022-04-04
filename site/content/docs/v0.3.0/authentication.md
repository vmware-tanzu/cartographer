# Authentication

## Owner Permissions

Cartographer requires a **service account** that permits all actions on the GVKs specified in **templates**.

### Per namespace service account

The operator provides a name for the service account that is used (but not the namespace). Typically, the operator will
ensure that a service account with sufficient privileges exists in each developer namespace.

The developer can still
[override the service account name](#developer-selects-the-name-of-a-service-account-in-their-namespace).

```yaml
---
kind: ClusterSupplyChain|ClusterDelivery
spec:
  serviceAccountRef:
    name: "operator-chosen-name"
    namespace: # not provided

---
kind: Workload|Deliverable
metadata:
  namespace: my-developer-ns
spec:
  serviceAccountName: # not provided
```

The selected service account is:

```yaml
---
kind: ServiceAccount
metadata:
  name: operator-chosen-name
  namespace: my-developer-ns
```

### Single service account

The operator provides a reference to a single service account that is used. The operator will ensure that one service
account with sufficient privileges exists.

The developer can still
[override the service account name](#developer-selects-the-name-of-a-service-account-in-their-namespace).

```yaml
---
kind: ClusterSupplyChain|ClusterDelivery
spec:
  serviceAccountRef:
    name: operator-chosen-name
    namespace: operator-chosen-namespace

---
kind: Workload|Deliverable
metadata:
  namespace: my-developer-ns
spec:
  serviceAccountName: # not provided
```

The selected service account is:

```yaml
---
kind: ServiceAccount
metadata:
  name: operator-chosen-name
  namespace: operator-chosen-namespace
```

### Developer selected service account

The developer provides a name for a service account that is in the same namespace as the owner (Workload/Deliverable)
they are creating. This takes precedence over operator provided service accounts. Of course the service account still
requires full permissions for the objects created by the blueprint.

```yaml
---
kind: ClusterSupplyChain|ClusterDelivery
spec:
  serviceAccountRef:
    name: # n/a
    namespace: # n/a

---
kind: Workload|Deliverable
metadata:
  namespace: my-developer-ns
spec:
  serviceAccountName: workload-specific-sa
```

The selected service account is:

```yaml
---
kind: ServiceAccount
metadata:
  name: workload-specific-sa
  namespace: my-developer-ns
```

### Default service account

If a service account is not specified in the blueprint or the owner, the `default` service account in the owner
namespace is used.

Note: The `default` service account is unlikely to have the necessary permissions.

```yaml
---
kind: ClusterSupplyChain|ClusterDelivery
spec:
  serviceAccountRef: {} # Not provided!

---
kind: Workload|Deliverable
metadata:
  namespace: my-developer-ns
spec:
  serviceAccountName: # Not provided!
```

The selected service account is:

```yaml
---
kind: ServiceAccount
metadata:
  name: default
  namespace: my-developer-ns
```

## Cartographer Controller Permissions

Cartographer has its own service account, `cartographer-controller` in the `cartographer-system` namespace. The
clusterrole that's bound to the service account is:

```bash
kubectl get clusterrole cartographer-controller-admin -oyaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cartographer-controller-admin
rules:
- apiGroups:
  - carto.run
  resources:
  - workloads/status
  - clustersupplychains/status
  - runnables/status
  - clusterdeliveries/status
  - deliverables/status
  verbs:
  - create
  - update
  - delete
  - patch
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - watch
  - get
  - list
```
