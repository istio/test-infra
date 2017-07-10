Infra
-----

This directory contains a Makefile and other resources for managing the Istio CI infrastructure. This infrastructure consists of a subset of the [k8s test-infra prow](https://github.com/kubernetes/test-infra/tree/master/prow) deployments.

### Managing a Cluster

The infrastructure runs on a k8s cluster. All of our tools are set up to handle running on GKE. The variables for project/zone/cluster are in the Makefile. The Makefile also contains commands for common management tasks.

#### Deployments

1. [hook](./cluster/hook_deployment.yaml)     - handle webhooks and create prow jobs
2. [plank](./cluster/plank_deployment.yaml)   - poll for prow jobs and start them, mark completed, statuses to Github
3. [deck](./cluster/deck_deployment.yaml)     - simple ui for prow jobs
4. [sinker](./cluster/sinker_deployment.yaml) - clean up old prow jobs
5. [tot](./cluster/tot_deployment.yaml)       - vendor build numbers (i.e., `<number>` in job name: `pilot-presubmit-<number>`)

We run the k8s-prow images. These images are built from source [here](https://github.com/kubernetes/test-infra/tree/master/prow).

### Creating a Job on Your Repo

#### Github Trigger

The most common pattern is to trigger a job on some sort of Github event, esp. on PRs and on PR merges. Prow has concepts for these two specific stages. The first, running jobs on a PR, is called a presubmit job. The second, running jobs after the PR is merged, is called a postsubmit.

Both of these types of jobs can be configured using the config configmap [here](./config.yaml). In the configmap, you are configuring on which repo to run a particular job, basic metadata like the name, and then the build image. For these to be triggered, you must add `trigger` to the list of plugins in the plugins configmap [here](./plugins.yaml). For example, to add a simple presubmit to `my-repo`, requires the following edits:

```yaml
# in config.yaml
triggers:
- repos:
  - istio/istio
  - istio/test-infra
  - istio/<my-repo> # ADD THIS LINE
# ...
presubmits:
  # ...
  istio/<my-repo>: # ADD THIS BLOCK
  - name: my-repo-presubmit
    context: prow/my-repo-presubmit.sh
    always_run: true
    rerun_command: "@istio-testing test this"
    trigger: "@istio-testing test this"
    branches:
    - master
    spec:
      containers:
      - image: gcr.io/istio-testing/prowbazel:0.1.2
    # ...

# in plugins.yaml
my-repo: # ADD THIS BLOCK
- trigger
```

##### Prow Bazel Build Image

The prowbazel build image [here](../docker/prowbazel) is preferred. Its entrypoint is a test harness that checks out the code at the appropriate ref, captures the logs and exit code, and writes these logs to a GCS bucket in a location and manner the k8s-test-infra UI, gubernator, can read.

In the repository of interest, add a `/prow/` directory. The test harness will look for `/prow/<job-name>.sh` and execute this script.

For example, a configuration that referenced the jobs `my-presubmit`, `my-postsubmit` and `my-job` would require a corresponding directory structure:

```bash
$ tree
.
├── anotherdir
│   └── ...
├── prow
│   ├── my-presubmit.sh
│   ├── my-postsubmit.sh
│   └── my-job.sh
├── LICENSE
└── README.md
```

Remember that when you add a job file, you need to set the execution bit!
```bash
$ chmod +x prow/my-repo-presubmit.sh
```

### Test-Infra Prow Jobs

This repository (istio/test-infra) also provides Prow jobs.

- `test-infra-presubmit` - run the linting and testing
