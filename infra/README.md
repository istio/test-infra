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
* GCS buckets for istio releases
* DNS configuration for istio.io

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

### `istio-prerelease-testing` GCP Project

This hosts `gcr.io/istio-prerelease-testing` and nothing else.

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

## GCS Buckets

Below catalogs all of the GCS buckets in one

| Bucket                                         | Project                  | Importance | Notes                                                                                          |
|------------------------------------------------|--------------------------|------------|------------------------------------------------------------------------------------------------|
| artifacts.istio-release.appspot.com            | istio-release            | Critical   | gcr.io/istio-release                                                                           |
| istio-release                                  | istio-io                 | Critical   | Production GCS artifacts of our releases                                                       |
| artifacts.istio-testing.appspot.com            | istio-testing            | Important  | gcr.io/istio-testing                                                                           |
| istio-build                                    | istio-testing            | Important  | Holds release artifacts for proxy, ztunnel, etc                                                |
| istio-prow                                     | istio-testing            | Important  | Stores all our test artifacts and results                                                      |
| istio-testgrid                                 | istio-testing            | Important  | Used by testgrid                                                                               |
| artifacts.istio-prerelease-testing.appspot.com | istio-prerelease-testing | Important  | gcr.io/istio-prerelease-testing                                                                |
| artifacts.istio-prow-build.appspot.com         | istio-prow-build         | Important  | gcr.io/istio-prow-build                                                                        |
| istio-private-build                            | istio-prow-build         | Important  | Private builds for proxy. Mirrors "istio-artifacts"                                            |
| istio-private-prerelease                       | istio-prow-build         | Important  | Private release-builder, mirrors istio-prerelease                                              |
| istio-prerelease                               | istio-io                 | Important  | Release builder publishes artifacts here before they are official released                     |
| istio-release-pipeline-data                    | istio-release            | Legacy     |                                                                                                |
| e2e-testing-log                                | istio-testing            | Legacy     |                                                                                                |
| istio-circleci                                 | istio-testing            | Legacy     |                                                                                                |
| istio-code-coverage                            | istio-testing            | Legacy     |                                                                                                |
| istio-flakes                                   | istio-testing            | Legacy     |                                                                                                |
| istio-gists                                    | istio-testing            | Legacy     |                                                                                                |
| istio-logs                                     | istio-testing            | Legacy     |                                                                                                |
| istio-presubmit-release-pipeline-data          | istio-testing            | Legacy     |                                                                                                |
| istio-stats                                    | istio-testing            | Legacy     | This was a oneoff upload                                                                       |
| istio-testing_cloudbuild                       | istio-testing            | Legacy     |                                                                                                |
| istio-tools                                    | istio-testing            | Legacy     |                                                                                                |
| fortio-data                                    | istio-io                 | Legacy     | Old fortio data                                                                                |
| artifacts.istio-io.appspot.com                 | istio-io                 | Legacy     | gcr.io/istio-io; only used for old artifacts                                                   |
| istio-io_cloudbuild                            | istio-io                 | Legacy     |                                                                                                |
| istio-artifacts                                | istio-testing            | Legacy?    | I believe formerly we used this for part of proxy release, now istio-build is used exclusively |
| istio-build-deps                               | istio-testing            | Legacy?    |                                                                                                |
| us.artifacts.istio-testing.appspot.com         | istio-testing            | Legacy?    | I am not sure what this is. Its part of GCR somehow                                            |
| istio-snippets                                 | istio-testing            | Legacy?    | Looks like it may have been used for istio.io tests, but likely not anymore                    |
| istio-private-artifacts                        | istio-prow-build         | Legacy?    | Private builds for proxy. Mirrors "istio-artifacts"                                            |
