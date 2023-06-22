# `gke-cluster` terraform module

This terraform module defines a GKE cluster following Istio conventions for prow clusters:

- GCP Service Account for nodes
- GKE cluster with some useful defaults
- No nodes are provided, they are expected to come from nodepools created via the [`gke-nodepool`] module
