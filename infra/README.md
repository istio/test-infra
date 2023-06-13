# Istio Infrastructure

This folder contains documentation and infrastructure-as-code for Istio's infrastructure.

Warning: this document is a work in progress and should not be seen as authoritative.
Istio is currently in the process of converting ad-hoc infrastructure to infrastructure-as-code;
during this time the definitions and documentation should be seen as best-effort only.

## Overview

Istio infrastructure spans a variety of platforms, but is primarily hosted on GCP.

### `istio-io` GCP Project

`istio-io` hosts:
* GCS bucket for terraform state (`istio-terraform`)

### `istio-release` GCP Project

`istio-release` hosts the `gcr.io/istio-release` registry, used in production.

### `istio-testing` GCP Project

`istio-testing` is a bit of a kitchen sync for various Istio testing efforts.

Most importantly, this project hosts:
* Prow control plane cluster
* gcr.io/istio-testing, used for hosting all our development builds (and testing tools)

### `istio-prow-build` GCP Project

This project contains our prow *build* clusters primary. This includes our public and private build infrastructure.
Additionally, private artifacts are stored in this project.

## Using terraform

Currently, terraform configuration is applied by humans.
Reach out to @howardjohn if changes are needed.

Within each package, the following commands can be used:

```shell
terraform fmt
terraform validate
terraform plan # Dry run to see what changes
terraform apply # Actually apply changes
```

Due to our current mixed state, generally its best to run `terraform plan` _before_ making any changes to detect any drift that may have occurred.

To apply changes, first submit a PR with the `terraform plan` output attached.
Once it is approved and merge, `terraform apply` can be run.

### Structure

Each project has its own folder, grouped by provider type.

For example:

```text
gcp
└── project-0
└── project-1
aws
└── project-0
```
