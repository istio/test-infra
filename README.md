# Istio Test Infra

This repository contains tools and configuration files for the testing and automation needs for Istio.

## Navigating this repository

* For information about interacting with Istio's CI/CD system (Prow), see [Working with Prow](https://github.com/istio/istio/wiki/Working-with-Prow).

* For information about _authoring_ new jobs, see [jobs/README.md](prow/config/jobs/README.md).

* For information about _operating_ the Prow control plane, see [prow/README.md](prow/README.md).

* For information about our cloud infrastructure, see [infra/README.md](infra/README.md).

* For information about tools we use...
  * [`automator`](tools/automator/README.md) helps automate dependency and other updates.
  * [`prowgen`](tools/prowgen/README.md) translates high level job definitions to ProwJob configurations.
  * [`prowtrans`](tools/prowtrans/README.md) translates ProwJob configurations to other ProwJob configurations. For example, to run the same jobs in the private infrastructure.
  * [`authentikos`](authentikos/README.md) writes access token to Kubernetes Secrets. This is not currently used by Istio, but is by other communities.
  * A variety of other tools live outside of this repo. Checkout [istio/tools](https://github.com/istio/tools) for more.
