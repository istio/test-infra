## update-e2e-cluster.sh

This script is used to update test clusters since the alpha-feature-cluster expires and is deleted in 30 days.
So we need to create new ones, and update secrets of kubeconfig in Prow to point to the new clusters. 

It always tries to create cluster with name `<repo>-e2e-rbac-rotation-1`, if gets error (if that cluster already exists),
tries to create `<repo>-e2e-rbac-rotation-2`. It always assumes there is only one of them exists, if not, exits with error.

* Flag
```
  -r  target repo
  -n  number of node
  -z  cluster zone
```

We need to update cluster for: auth, broker, pilot, mixer, istio
Example:
```Bash
$ ./scripts/update-e2e-cluster.sh -r pilot -n 10
$ ./scripts/update-e2e-cluster.sh -r auth -n 4
```

## cleanup-cache

This script is used to clean up bazel cache in CI cluster.

- Cleanup Jenkins cluster:
```Bash
$ ./scripts/cleanup-cache
```

- Cleanup Prow cluster:
```Bash
$ scripts/cleanup-cache -c prow -z us-west1-a -s cloud.google.com/gke-nodepool=build-pool
```
