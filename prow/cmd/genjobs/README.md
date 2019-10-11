# Genjobs

## Description

`genjobs` is a CLI tool used to generate **private** Github jobs from existing Prow job definitions. It translates *existing* jobs by adding decoration to the Prow job spec.

## Installation

```console
$ go get -u istio.io/test-infra/prow/cmd/genjobs
```

## Usage

The following is a list of supported options for `genjobs`. The only **required** option is `-m, --mapping`, which is the translation mapping between public/private Github organizations.

```shell
      --branches strings         Branches to generate job(s) for.
      --bucket string            GCS bucket name to upload logs and build artifacts to. (default "istio-private-build")
      --clean                    Clean output directory before job(s) generation.
      --cluster string           GCP cluster to run the job(s) in. (default "private")
  -i, --input string             Input directory containing job(s) to convert. (default ".")
      --job-blacklist strings    Jos(s) to blacklist in generation process.
      --job-whitelist strings    Job(s) to whitelist in generation process.
  -l, --labels stringToString    Prow labels to apply to the job(s). (default [preset-service-account=true])
  -m, --mapping stringToString   Mapping between public and private Github organization(s). (default [])
  -o, --output string            Output directory to write generated job(s). (default ".")
  -b, --repo-blacklist strings   Repositories to blacklist in generation process.
  -w, --repo-whitelist strings   Repositories to whitelist in generation process.
      --ssh-key-secret string    GKE cluster secrets containing the Github ssh private key. (default "ssh-key-secret")
```

## Example

Translate all public jobs with `istio` organization to private jobs with `istio-private` organization in `./jobs` directory:

```console
$ genjobs --mapping istio=istio-private --input ./jobs --output ./jobs
```

Limit job generation to *specific* branches:

```console
$ genjobs --mapping istio=istio-private --branches master
```

Limit job generation to *specific* repositories:

```console
$ genjobs --mapping istio=istio-private --repo-whitelist cni, api
```

Limit job generation to *specific* job names:

```console
$ genjobs --mapping istio=istio-private --job-whitelist build_bots_postsubmit
```

Define the `bucket` to upload job results to:

```console
$ genjobs --mapping istio=istio-private --bucket istio-private-build
```

Define the `ssh-key-secret` secret to authorize repository clone with:

```console
$ genjobs --mapping istio=istio-private --ssh-key-secret ssh-key-secret
```

Add additional `labels` to the job:

```console
$ genjobs --mapping istio=istio-private --labels preset-service-account=true
```

Set the `cluster` on which the jobs will run:

```console
$ genjobs --mapping istio=istio-private --cluster private
```

Delete jobs in destination path prior to generation:

```console
$ genjobs --mapping istio=istio-private --clean
```
