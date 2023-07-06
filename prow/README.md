# Prow

See [upstream prow](https://github.com/kubernetes/test-infra/tree/master/prow) documentation for more detailed and generic information about what prow is and how it works.

## Upgrading Prow

Prow is automatically upgraded on a regular cadence by the `ci-prow-autobump` job.

## Prow Secrets

Some of the prow secrets are managed by kubernetes external secrets, which
allows prow cluster creating secrets based on values from google secret manager
(Not necessarily the same GCP project where prow is located). See more detailed
instruction at [Prow Secret](https://github.com/kubernetes/test-infra/blob/master/prow/prow_secrets.md).
