# `workload-identity-serviceaccount` terraform module

This terraform module defines a GCP service account intended solely for use
by pods running in GKE clusters in a given project, running as a given K8s
service account in a given namespace.

This is forked (mostly as-is) from https://github.com/kubernetes/k8s.io/blob/main/infra/gcp/terraform/modules/workload-identity-service-account/README.md.
