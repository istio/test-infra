# Prow

See [upstream prow](https://github.com/kubernetes/test-infra/tree/master/prow) documentation for more detailed and generic information about what prow is and how it works.

## Upgrading Prow

Please check recent [prow announcements](https://github.com/kubernetes/test-infra/tree/master/prow#announcements) before updating, if you are not already familiar with them.

```bash
$ prow/bump.sh --auto
```

- Commit the change and merge the PR.
- Deployment will occur as a postsubmit job
- Watch pods (below) and watch for problems
- prow.istio.io and watch for problems
- Look at stack driver logs (go/istio-prow-debug) and look for problems

## Watch pods

```bash
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

## Check logs

```bash
kubectl logs -l app=deck # or the appropriate label like app=hook
# or a specific pod: kubectl logs deck-3621325446-00drl
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
- `test-infra-cleanup-GKE` - Delete GKE clusters left behind in test environment due to jobs being killed inproperly.

### Manually Trigger a Prow Job

```bash
# Assuming you cannot click the rerun button on prow.istio.io,
# and if you are oncall, do the following:
go get -u k8s.io/test-infra/prow/cmd/mkpj

mkpj --job=FOO > ~/foo.yaml # and answer interactive questions

# Contact #oncall on slack to ensure you are approved to do the following:
kubectl --context=istio-prow create -f ~/foo.yaml
```

### Update Config File

File `rewriteConfig.go` rewrites config file with new branches to be
added under the field of `repos:` for specific repos based on the config
content already existing in `master` branch. Content in new branch would
be identical to the content in `master` branch with additional lines to
restrict merge blocks to admin approval if it is not already present in
the `master` content.

The file requires the following flags:

-`InputFileName` - The name and path to the config file that requires to be rewritten.
-`NewBranchName` - Name of new branch to be added to the config file.
-`ReposeToAdd` - Names of repos the new branch should be added to separated by ','. Default to be "proxy,istio,istio-releases".

The file should be run with `go build`, `go test` (for test file rewriteConfig_test.go) and `go run`.

For an original section of config.yaml

```yaml
repos:
  istio:
    branches:
        <<: *blocked_branches
        master:
          protect: true
```

when `InputFileName=config.yaml`, `NewBranchName=newBranch` and `ReposeToAdd=istio`, the result of adding new branch to the original section would be:

```yaml
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

## Prow Secrets

Some of the prow secrets are managed by kubernetes external secrets, which
allows prow cluster creating secrets based on values from google secret manager
(Not necessarily the same GCP project where prow is located). See more detailed
instruction at [Prow Secret](https://github.com/kubernetes/test-infra/blob/master/prow/prow_secrets.md).
