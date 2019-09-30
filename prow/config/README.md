# Istio Prow Job Configuration

This package defines and generates all prow jobs that will be run in pre/postsubmit.

Jobs reside in [jobs/](./jobs/). From there, [`generate.go`](./generate.go) will process the config and turn them into a valid [prow job](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md).

This generation layer simplifies the configuration and allows use to provide opinionated defaults.

## Job Syntax

Before is an example, annotated job config

```yaml
# REQUIRED. Defines what repo these jobs should run for
repo: istio/istio

# Defines what branches to run these jobs for. Multiple can be provided
# The branch name will be appended to the job name (e.g tests -> tests-master)
branches:
  - master

# Defines the actual jobs
jobs:
  # A basic test requires just a name and a command to run
  - name: unit-tests
    command: [make, test]
  - name: integration-tests
    # type defines when the job will run. Valid options are [presubmit, postsubmit].
    # by default a presubmit and postsubmit job will be created with the same config
    type: postsubmit
    # resources determines what resource requests and limits to use. See the resources section below
    resources: large
    command: [prow/istio-lint.sh]
    # requirements specify what dependencies a test has. Valid options are:
    # - root, which will give the test a privileged container. Note: currently this is the default but will change in the future
    # - gcp, which will give the test access to GCP secrets. This is needed for pushing to GCR or using Boskos
    # - kind, which will configure the test to allow kind (https://kind.sigs.k8s.io) to run
    requirements: [gcp]
  - name: hello-world
    command: [echo, "hello world"]
    # modifiers change various parts of the test config. See the values below
    modifiers:
    - skipped # if set, the test will run only in postsubmit or by explicitly calling /test on it
    - hidden # if set, the test will run but not be reported to the GitHub UI
    - optional # if set, the test will not be required

# Defines preset resource allocations for tests
# If a job doesn't specify one, the "default" will be used
resources:
  default:
    requests:
      memory: "3Gi"
      cpu: "3000m"
    limits:
      memory: "24Gi"
      cpu: "3000m"
  # Define another preset, "large", with higher allocations
  large:
    requests:
      memory: "16Gi"
      cpu: "3000m"
    limits:
      memory: "24Gi"
      cpu: "3000m"
```

## Generating the config

The config generate has a few commands that can be run with:

```bash
$ go run generate.go [diff|print|write|check]`
```

* diff will produce a semantic diff of the current config and the newly generated config. This is useful when making changes
* print will print out all generated config to stdout
* write will write out generated config to the appropriate job file
* check will strictly compare the generated config to the current config, and fail if there are any differences. This is useful for a CI gate to ensure config is up to date
