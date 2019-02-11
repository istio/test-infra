Prow Bazel Image
----------------

This image is suitable for projects that use bazel to build and test code. The image comes with bazel, go, docker, and gcloud utilities.

### Building Image

Use `make image` to build the docker image. Run `make push` to send it to the istio-testing gcr registry.

### Running Image Locally

Use `make run` to start the docker image locally.

### Dependency on `bootstrap.py`

This `bootstrap.py` file is copied here so that we have a local copy. It is copied from the [k8s test-infra version](https://github.com/kubernetes/test-infra/tree/master/jenkins).

We modify bootstrap.py in two ways:

1. It looks for job scripts in `<source-repo>/jobs/<script>.sh` rather than in the kubernetes `test-infra/jobs/<script>.sh`.
2. It does _not_ set up the "magic" environment.

Long term, if we use `bootstrap.py`, it would be prudent to push these changes upstream, perhaps as flags to further configure bootstrap.py.

Our dependency on this script is because it appropariately writes test job results to a GCS bucket in such a way that gubernator (a basic UI k8s test-infra run to view job results) understands. The specification of this structure is inherent in bootstrap.py and the gubernator front-end.

### Upgrade log

* 0.3.3: golang 1.8 -> 1.9
* 0.3.4: add protoc 3.5.0
* 0.3.5: add protoc 3.5.0 include files
* 0.3.6: add fpm and its ruby dependencies
* 0.4.0: update bazel to 0.10.0
* 0.4.1: add go-junit-report
* 0.4.2: add /opt/go/bin to PATH
* 0.4.3: update bazel to 0.11.0
* 0.4.4: update golang to 1.10
* 0.4.5: add clang 5.0
* 0.4.6: update go-junit-report to handle concurrent execs
* 0.4.7: update bazel to 0.11.1
* 0.4.8: update golang to 1.10.1
* 0.4.9: adds helm 2.7.2
* 0.4.10: adds ninja and updates bazel to 0.15.2
* 0.4.11: update clang to 6.0
* 0.4.12: update golang to 1.10.4 and docker to 18.06.1-CE
* 0.4.13: Update docker to use overlay2 fs for performance and bazel to 0.16.1
* 0.5.0: Update bazel to 0.22 and clang to 7
* 0.5.1: Add lld and libc++
