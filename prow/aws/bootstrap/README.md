# AWS Prow bootstrap manifests

These manifests establish the Kubernetes authorization boundary used by the
`prow-deployer` workload. They are intentionally outside the directories
applied by the regular Prow deployment job.

Apply the Terraform configuration first so the Pod Identity role and EKS
access entries exist. Then apply the namespace and RBAC manifests with
cluster-admin credentials:

```bash
kubectl --context prow apply \
  -f ../cluster/test_pod_namespace.yaml \
  -f prow/prow-deployer_rbac.yaml

kubectl --context prow-build apply \
  -f ../cluster/build/test_pod_namespace.yaml \
  -f prow-build/prow-deployer_rbac.yaml

kubectl --context prow-private apply \
  -f ../cluster/private/test_pod_namespace.yaml \
  -f prow-private/prow-deployer_rbac.yaml
```

Run these commands from the `prow/aws/bootstrap` directory. Reapply the
manifests as an administrator whenever the deployer permissions change.
