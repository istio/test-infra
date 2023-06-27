# Istio Job Configuration

This folder providers Istio test/job definitions.
Istio uses [Prow](https://docs.prow.k8s.io/docs/) for all tests.

However, we use an Istio-specific higher level configuration for jobs to help abstract away some of the details of writing jobs.
This low level details of this can be found in the [`prowgen` README](../../../tools/prowgen/README.md), but this is a bit
dense - `prowgen` is used by other projects, so its scope is beyond what most Istio job authors need to understand.

## Adding a job

Each repository is configured with its own YAML file.
By convention, this follows the repository name, such as `api.yaml`.
Additionally, you'll find copies of these for older release branches (`api-1.0.yaml`) - usually these are not modified once the branch is cut.

Adding or modifying a job is as simple as adding an entry to the `jobs` field.
A trivial job would look like:

```yaml
- name: build
  command: [make, build]
```

### Job Schema

All the supported fields for a job can be checked from [spec.go](./pkg/spec/spec.go), but below covers the important ones.

* `name`: all jobs have a name. Keep it simple. The name will automatically have the repo name, job type, and branch name appended, so there is no need to include these. `unit-tests` is an example of a good name.
* `command`: what command to run.
    The test image has a binary, `entrypoint`, that configures various things such as IPv6 and docker, making this a command part of the `command`; tests will check this is set where required.
    Generally, its best to keep the command simple and point to a script or make target.
    This allows testing changes by modifying the scripts, and keeps things encapsulated in the repository.
* `types`: this defines the type of job. This can be any combination of `periodic`, `presubmit`, and `postsubmit`.
   By default, jobs will run in `presubmit` and `postsubmit`, which is generally recommended.
* `modifiers`: change properties of the job. Possible values are:
    * `presubmit_skipped`: the test will only be run in presubmit by explicitly calling `/test` on it
    * `presubmit_optional`: the test will not be required in presubmit
    * `hidden`: the test will run but not be reported to the GitHub UI
* `requirements`: these act as modifiers on the job, giving the job access to different resources.
    The full list can be found in [`.base.yaml`](.base.yaml), but the most common are:
    * `kind`: Sets up the job to be able to access a [`kind`](https://kind.sigs.k8s.io/) cluster. Jobs testing against Kubernetes will need this.
    * `docker`: Sets up the job to be able to access docker.
    * `cache`, `gocache`, `cratescache`: Sets up caching for Go Modules, Go Builds, and Crates (Rust), respectively.
        Caches are ephemeral mounts on the host, and not shared between hosts, so they are fairly low hit rate.
* `resources`: sets the compute resources requested by the job. References presets defined in [Repo Schema](#repo-schema).
* `architectures`: sets the architectures to run the job as. Defaults to `[amd64]`, allows `arm64` and `amd64`.

### Repo Schema

Along with the main `jobs` field, discussed in [Jobs Schema](#job-schema), there is some repo-level configuration as well

* `org`/`repo`: the GitHub repository. For example, `istio` and `test-infra`.
* `support_release_branching`: should be `true` if the repo uses release branches
* `image`: image to run tests under. This should usually be `gcr.io/istio-testing/build-tools`; if it is, automation updates the version.
* `resources_presets`: configure resources presets to reference using `resources` in the job. `default` is special.

### Testing Changes

The `TestJobs` Go test runs a variety of checks against tests to make sure they meet our consistency and security guidelines.
These can be run with `make test`.

## Test environment

Jobs run as Kubernetes `Pods`, running in GKE clusters.

Most jobs run with the [`build-tools`](https://github.com/istio/tools/blob/master/docker/build-tools/Dockerfile) image, which
has all the tools Istio uses installed.
An optional [`entrypoint`](https://github.com/istio/tools/blob/master/docker/build-tools/prow-entrypoint.sh) command is available in this container, which sets up docker.

Prow augments our pod with a variety of helper containers, that do things like upload logs and artifacts.
Any files in `$ARTIFACTS` will be persistently uploaded when the job completes.
A variety of [environment variables](https://docs.prow.k8s.io/docs/jobs/#job-environment-variables) are also injected.
