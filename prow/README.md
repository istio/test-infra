# Infra

This directory contains a Makefile and other resources for managing the Istio CI infrastructure. This infrastructure consists of a subset of the [k8s test-infra prow](https://github.com/kubernetes/test-infra/tree/master/prow) deployments.

## Managing a Cluster

The infrastructure runs on a k8s cluster. All of our tools make it easy to run on GKE, although another k8s provider could be used. The variables for the GCP project/zone/cluster are in the Makefile. The Makefile also contains commands for common management tasks.

### Deployments

1. [hook](./cluster/hook_deployment.yaml)               - handle webhooks and create prow jobs
2. [plank](./cluster/plank_deployment.yaml)             - poll for prow jobs and start them, mark completed, statuses to Github
3. [deck](./cluster/deck_deployment.yaml)               - simple ui for prow jobs
4. [sinker](./cluster/sinker_deployment.yaml)           - clean up old prow jobs
5. [tot](./cluster/tot_deployment.yaml)                 - vendor build numbers (i.e., `<number>` in job name: `pilot-presubmit-<number>`)
6. [horologium](./cluster/horologium_deployment.yaml)   - start periodic jobs

We run the k8s-prow images. These images are built from source [here](https://github.com/kubernetes/test-infra/tree/master/prow).

### Upgrading Prow

Please check Prow announcements before starting an upgrade (https://github.com/kubernetes/test-infra/tree/master/prow#announcements)

It is a good idea to take a quick glance at the [Kubernetes Prow config](https://github.com/kubernetes/test-infra/blob/master/prow/config.yaml)
if you see anything new that looks backward incompatible.
There should be no breaking changes, but at the time of this writing (Aug. ‘17) the project is still somewhat in flux

* Check the current versions of the Prow images [here](https://github.com/kubernetes/test-infra/tree/master/prow/cluster).
Compare these with the version in our [deployment files](https://github.com/istio/test-infra/tree/master/prow/cluster) and update with more recent setting. Before updating a flag make sure it is required, and check the default.

* Before starting upgrade Run the following command in another terminal

```
$ watch kubectl get pods

Every 2.0s: kubectl get pods                                                                                                         Fri Aug 11 15:40:31 2017

NAME                         READY     STATUS    RESTARTS   AGE
deck-3621325446-00drl        1/1       Running   0          54m
deck-3621325446-9pdqw        1/1       Running   0          55m
deck-3621325446-njnwk        1/1       Running   0          54m
hook-3348033068-2tdd3        1/1       Running   0          45m
hook-3348033068-x99bf        1/1       Running   0          45m
horologium-617344823-js4mk   1/1       Running   0          50m
plank-302445171-92rfx        1/1       Running   0          41m
sinker-799599164-z44wj       1/1       Running   0          34m
tot-763621987-pktpj          1/1       Running   0          37m

```

* In case of deployment failure, check the logs by running

```
$ kubectl logs deck-3621325446-9pdqw
```

* You might notice some issues related to config error. In that case fix and redeploy the config

```
$ make update-config
```

* In some other case, flags or volume might be missing, make sure you update your config properly.
Once done re-run the deployment issue that failed.


* Start by updating deck. This deployment has multiple replicas, so you can resolve the issues as you see them.

```
$ make deck-deployment
```

* Deploy horologium.

```
$ make horologium-deployment
```

* Deploy hook. This deployment also has multiple replicas, so failure will not impact workflows.

```
$ make hook-deployment
```

Update to the following are asynchronous
```
$ make plank-deployment
$ make tot-deployment
$ make sinker-deployment
```

### Clearing Up Prow State

If prow is an unrecoverable state. We can reset Prow state by doing the
following:

```
$ make stopd
$ kubectl delete prowjobs --all
$ kubectl delete pods -n test-pods --all
$ make startd
```

## Creating a Job on Your Repo

### Github Trigger

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
      - image: gcr.io/istio-testing/prowbazel:0.4.11
    # ...

# in plugins.yaml
my-repo: # ADD THIS BLOCK
- trigger
```

### Upload artifacts after job finishes

Since the pod is gone as soon as the test finishes, it's hard to track back to see what happened. Prow is able to upload temp files, logs amd any other artifacts to cloud storage. What you need to do is put aritifact into `_artifacts` directory.

`_artifacts` is created by Prow in root directory of current project. Take istio/istio as a example:

Test harness will checkout code to directory `$GOPATH/src/github.com/istio/istio`

But in most of our tests we create a soft link between `$GOPATH/src/github.com/istio/<repo>` and `$GOPATH/src/istio.io/<repo>` since we import from `istio.io`

So `_artifacts` dir will be generated before running tests and accessable at both `$GOPATH/src/github.com/istio/istio/_artifacts` and `${GOPATH}/src/istio.io/istio/_artifacts`

Then you can access to the artifact through "artifacts" at [gubernator](https://k8s-gubernator.appspot.com/build/istio-prow/pull/istio_istio/1025/e2e-suite-rbac-no_auth/1006/)


### Prow Bazel Build Image

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
$ chmod +x prow/my-presubmit.sh
```

### Test-Infra Prow Jobs

This repository (istio/test-infra) also provides Prow jobs.

- `test-infra-presubmit` - Run the linting and testing
- `test-infra-update-deps` - Create automatically dependency update PRs in target repos.
- `test-infra-cleanup-GKE` - Delete GKE clusters left behind in test environment due to jobs being killed inproperly.
- `test-infra-cleanup-cluster` - Delete k8s namespace left behind in testing cluster due to tests finishing inproperly. 

### Manually Trigger a Prow Job


> **Never do this unless necessary AND you are authorized by istio EngProd team (istio.slack.com test-infra channel)**


Connect to Prow cluster and then:
```bash
$ git clone https://github.com/kubernetes/test-infra
$ bazel build //prow/cmd/mkpj
```
And then specify the script and the job you want to trigger:
``` bash
$ bazel-bin/prow/cmd/mkpj/mkpj --config-path ~/istio/test-infra/prow/config.yaml --job test-infra-presubmit | kubectl create -f -
```

And give the required information:
```
Base ref (e.g. master): master
Base SHA (e.g. 72bcb5d80): 6419408170738a60cf04f963e4ae139028bf0b5b
PR Number: 465                                     
PR author: yutongz
PR SHA (e.g. 72bcb5d80): d7e1ef38cf294de11062a6760d073827585af219
```
Prow should respond:
```
prowjob "e810668f-9435-11e7-9a4d-784f43915c4d" created
```

### Update Config File

File `rewriteConfig.go` rewrites config file with new branches to be added under the field of `repos:` for specific repos based on the config content already existing in `master` branch. Content in new branch would be identical to the content in `master` branch with additional lines to restrict merge blocks to admin approval if it is not already present in the `master` content. 

The file requires the following flags:

-`InputFileName` - The name and path to the config file that requires to be rewritten.
-`NewBranchName` - Name of new branch to be added to the config file.
-`ReposeToAdd` - Names of repos the new branch should be added to separated by ','. Default to be "proxy,istio,istio-releases".

The file should be run with `go build`, `go test` (for test file rewriteConfig_test.go) and `go run`.

For an original section of config.yaml

```
repos:
  istio:
    branches:
        <<: *blocked_branches
        master:
          protect: true
```
when `InputFileName=config.yaml`, `NewBranchName=newBranch` and `ReposeToAdd=istio`, the result of adding new branch to the original section would be:

```
repos:
  istio:
    branches:
        <<: *blocked_branches
        master:
          protect: true
        newBranch:
          protect: true
          required_status_checks:
            contexts:
            - "merges-blocked-needs-admin"
```
