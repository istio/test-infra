# Genjobs

## Description

`genjobs` is a command-line interface (CLI) tool used to generate [Prow job](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md)
objects by transforming existing Prow job objects.

## Installation

Install using Golang:

```shell
GO111MODULE="on" go get -u istio.io/test-infra/prow/genjobs
```

Install using Docker:

```shell
docker pull gcr.io/istio-testing/genjobs:latest
```

Install from source:

```shell
git clone --depth=1 https://github.com/istio/test-infra.git
cd ./test-infra/prow/genjobs/
go install
```

## Usage

Run using Golang:
> Ensure `$GOPATH/bin` is on your `$PATH`; or execute `$GOPATH/bin/genjobs` directly.

```shell
genjobs <options>
```

Run using Docker:

```shell
docker run gcr.io/istio-testing/genjobs:latest <options>
```

The following is a list of supported options for `genjobs`. The only **required** option is `-m, --mapping`, which is the translation mapping between public/private Github organizations.

```console
  -a, --annotations stringToString   Annotations to apply to the job(s) (default [])
      --branches strings             Branch(es) to generate job(s) for.
      --branches-out strings         Override output branch(es) for generated job(s).
      --bucket string                GCS bucket name to upload logs and build artifacts to.
      --channel string               Slack channel to report job status notifications to.
      --clean                        Clean output files before job(s) generation.
      --cluster string               GCP cluster to run the job(s) in.
      --configs strings              Path to files or directories containing yaml job transforms.
      --dry-run                      Run in dry run mode.
  -e, --env stringToString           Environment variables to set for the job(s). (default [])
      --global string                Path to file containing global defaults configuration.
  -i, --input string                 Input file or directory containing job(s) to convert. (default ".")
      --job-blacklist strings        Job(s) to blacklist in generation process.
  -t, --job-type strings             Job type(s) to process (e.g. presubmit, postsubmit. periodic). (default [presubmit,postsubmit,periodic])
      --job-whitelist strings        Job(s) to whitelist in generation process.
  -l, --labels stringToString        Prow labels to apply to the job(s). (default [])
  -m, --mapping stringToString       Mapping between public and private Github organization(s). (default [])
      --modifier string              Modifier to apply to generated file and job name(s). (default "private")
  -o, --output string                Output file or directory to write generated job(s). (default ".")
      --override-selector            The existing node selector will be overridden rather than added to.
  -p, --presets strings              Path to file(s) containing additional presets.
      --refs                         Apply translation to all extra refs regardless of repo.
  -b, --repo-blacklist strings       Repositories to blacklist in generation process.
  -w, --repo-whitelist strings       Repositories to whitelist in generation process.
      --rerun-orgs strings           GitHub organizations to authorize job rerun for.
      --rerun-users strings          GitHub user to authorize job rerun for.
      --resolve                      Resolve and expand values for presets in generated job(s).
      --selector stringToString      Node selector(s) to constrain job(s). (default [])
  -s, --sort string                  Sort the job(s) by name: (e.g. (asc)ending, (desc)ending).
      --ssh-clone                    Enable a clone of the git repository over ssh.
      --ssh-key-secret string        GKE cluster secrets containing the Github ssh private key.
      --verbose                      Enable verbose output.
```

## Example

Transform all public jobs with `istio` organization to private jobs with `istio-private` organization in `./jobs` directory:

```shell
genjobs --mapping istio=istio-private --input ./jobs --output ./jobs
```

To perform the same transforms using a yaml configuration file `./config.yaml`:

```yaml
# config.yaml

transforms:
- mapping:
    istio: istio-private
  input: ./jobs
  output: ./jobs
```

```shell
genjobs --configs=./config.yaml
```

Limit job generation to *specific* branches:

```shell
genjobs --mapping istio=istio-private --branches master
```

Limit job generation to *specific* repositories:

```shell
genjobs --mapping istio=istio-private --repo-whitelist cni, api
```

Limit job generation to *specific* job names:

```shell
genjobs --mapping istio=istio-private --job-whitelist build_bots_postsubmit
```

Define the `bucket` to upload job results to:

```shell
genjobs --mapping istio=istio-private --bucket istio-private-build
```

Define the `ssh-key-secret` secret to authorize repository clone with:

```shell
genjobs --mapping istio=istio-private --ssh-key-secret ssh-key-secret
```

Add additional `labels` to the job:

```shell
genjobs --mapping istio=istio-private --labels preset-service-account=true
```

Set the `cluster` on which the jobs will run:

```shell
genjobs --mapping istio=istio-private --cluster private
```

Delete jobs in destination path prior to generation:

```shell
genjobs --mapping istio=istio-private --clean
```

## Changelog

- 0.0.1: initial release
- 0.0.2: add `--branches-out` option for overriding the output branch(es) of generated jobs.
- 0.0.3: add `--verbose` option to enable verbose output and `--configs` option for specifying transforms via a yaml configuration file(s).
- 0.0.4: add `defaults` key for specifying _file-level_ defaults, support a `.defaults.yaml` file for _local_ defaults, and add `--global` option for _global_ defaults.
- 0.0.5: rename `--extra-refs` option to `--refs` and designate `extra-refs` key for specifying a list of extra refs to append to job.
