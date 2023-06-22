# `gke-nodepool` terraform module

This terraform module defines a GKE nodepool following Istio conventions, including:

- Workload Identity is enabled by default for this nodepool
- Legacy metadata endpoints are disabled
- Auto-repair and auto-upgrade are enabled

It is assumed that the associated GKE cluster has been provisioned using the `gke-cluster` module
