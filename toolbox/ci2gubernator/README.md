# Integrating Non-Prow CI with Testgrid, Gubernator, and Flakiness Metrics

Istio now uses both Prow and CircleCI for pre-submit, post-submit, and periodic jobs.
Each CI stores test results, logs, and artifacts in different formats and at
different locations. Many downstream data processing systems such as
[Testgrid](https://k8s-testgrid.appspot.com/istio-presubmits)
require the results to be in a well defined structure, which Prow conforms to but
CircleCI does not. To expand the coverage to any non-prow CI such as CircleCI (in
case that istio moves to other managed solution in the future), a
CI-agnostic adaptor is needed to parse the results, put them in the right file
structures, and upload them to the right location.


## Introduction

Gubernator and Testgrid are both frontends that display test results stored in GCS.

To correctly interpret jobs results, they expect that any job directory on GCS is formulated as the following.

```
.
├── artifacts         # all artifacts must be placed under this directory
│   └── junit_*.xml   # JUnit XML reports from the build
├── build-log.txt     # std{out,err} from the build
├── finished.json     # metadata uploaded once the build finishes
└── started.json      # metadata uploaded once the build starts
```

Every run should upload `started.json`, `finished.json`, and `build-log.txt`, and can optionally upload JUnit XML and/or other files to the `artifacts/` directory.

The following fields in `started.json` are honored:

```json
{
    "timestamp": "seconds after UNIX epoch that the build started",
    "pull": "$PULL_REFS from the run",
    "repos": {
        "org/repo": "git version of the repo used in the test",
    },
}
```

The following fields in `finished.json` are honored:

```json
{
    "timestamp": "seconds after UNIX epoch that the build finished",
    "result": "SUCCESS or FAILURE, the result of the build",
    "metadata": "dictionary of additional key-value pairs that will be displayed to the user",
}
```

Any artifacts from the build should be placed under `./artifacts/`. Any jUnit
XML reports should be named `junit_*.xml` and placed under `./artifacts` as well.

The binary `ci_to_gubernator` constructs these files and uploads these artifacts along with the log to the right location on GCS.


## Usage

Every call to `ci_to_gubernator` needs authentication in order to upload artifacts.
It requires a GCP service account json that includes the API token, which can be
supplied using the `--service_account` to specify the path the service account json.

At the start of a build, execute the following command to create and upload `started.json` on GCS.
If the job is running as part of the presubmit, one should
also specify the `--stage=presubmit` flag. Presubmits are uploaded to a different
locations on GCS so we could have different panels on testgrid that makes
multiplexing results attainable.

```bash
$ ci_to_gubernator --job_starts \
	--sha=<PULL_REFS> \
	--org=<GITHUB_ORG> \
	--repo=<GITHUB_REPO> \
	--job=<CI_JOB_NAME> \
	--build_number=<CI_BUILD_NUMBER> \
	--pr_number=<GITHUB_PULL_REQUEST_NUMBER>
```

At the end of a build, execute the following command to
* create `finished.json`
* upload `finished.json`, `artifacts/junit_*.xml`, and `build-log.txt` to GCS
* update `latest-build.txt` of this job with the current build number

The same rules on `--stage=presubmit` usage apply in here.

```bash
$ ci_to_gubernator \
	--exit_code=<BUILD_PROCESS_EXIT_STATUS> \
	--sha=<PULL_REFS> \
	--org=<GITHUB_ORG> \
	--repo=<GITHUB_REPO> \
	--job=<CI_JOB_NAME> \
	--build_number=<CI_BUILD_NUMBER> \
	--build_log_txt=<PATH_TO_LOG_FILE>
```

Notice that current running jobs are going to update `latest-build.txt`, which is not safe without synchronization. To serialize the access to `latest-build.txt`, we use [gsclock](https://github.com/marcacohen/gcslock) that takes advange of [object versioning](https://cloud.google.com/storage/docs/object-versioning) of GCS.

## Existing Integration with CircleCI

Examples of using `ci_to_gubernator` can be found at [CircleCI config](https://github.com/istio/istio/blob/master/.circleci/config.yml#L39-63) and its [helper script `ci2gubernator.sh`](https://github.com/istio/istio/blob/master/bin/ci2gubernator.sh).

`ci2gubernator.sh` encapsulates `ci_to_gubernator` with other logics around
CircleCI, such as constructing command-line arguments to `ci_to_gubernator` using
environment variables provisioned by Circle CI, and decrypting the encrypted service
account json file.

Orchestration of steps are done in CircleCI config yaml. An e2e job could be abstract into the following:

```yaml
e2e-example:
  <<: *integrationDefaults
  steps:
    - <<: *initWorkingDir
    - checkout
    - attach_workspace:
        at:  /go
    - <<: *markJobStartsOnGCS
    ...
    ... calling make targets to run tests ...
    ...
    - <<: *recordZeroExitCodeIfTestPassed
    - <<: *recordNonzeroExitCodeIfTestFailed
    - <<: *markJobFinishesOnGCS
    - store_artifacts:
        path: ...
```

To make proper use to `ci2gubernator.sh`, the four steps of `markJobStartsOnGCS`,
`recordZeroExitCodeIfTestPassed`, `recordNonzeroExitCodeIfTestFailed`, and
`markJobFinishesOnGCS` indispensable.

We need at least two separate calls to `ci2gubernator.sh` at the start
(`markJobStartsOnGCS`) and the end (`markJobFinishesOnGCS`) of the job. The
remaining two steps are mutually exclusive (exactly one or the other will execute)
and are used to record the exit code of the make target, assuming the make target
is the last command that was executed before either of these two. They rely on the
`on_success` and `on_fail` conditions to understand the exit code of the test target
and writes the exit code to a temporary file, so the later step
`markJobFinishesOnGCS` knows how to proceed with the results.

### Add A New Job to CircleCI and Testgrid

For a new job just added to CircleCI config, following the steps we just laid out
also automatically creates under the right bucket a new folder named after the new
job to store all of the build results.

Yet there is an additional step to add this new job to Testgrid. One has to update
the [`config.yaml`](https://github.com/kubernetes/test-infra/blob/master/testgrid/config.yaml#L2476)
used by Testgird by adding a new test group with GCS prefix to results, adding that
test group to the right dashboard, and set up failure threshold for alert and the
oncall mailing list. The [Testgrid README](https://github.com/kubernetes/test-infra/blob/master/testgrid/README.md)
has all the detailed steps on how to do so.

After your PR to update Testgrid config is merged, the Testgrid webpage should
reflect your change in matters of minutes. The actual test-case level visibility
though might takes a bit longer than that to be up. If the issue persists for more
than half a day, please reach out to the kubernetes team (@BenTheElder @fejta
@krzyzacy and @ixdy are all nice people to talk to).

The same steps applies when a job becomes stale and one were to remove it from
Testgrid. All in all, the job list defined by CI is manually propagated to Testgrid
as of right now. We are looking at automation in the near future.

## Trouble Shoots

#### All test cases passed but job status reported failure

Testgrid does no more than rendering the junit reports across multiple runs of a
given job. A junit report is an XML file that summarizes an execution of a job. An example junit report is the following.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite tests="3" failures="0" time="144.101" name="istio.io/istio/tests/e2e/tests/simple">
    <properties>
      <property name="go.version" value="devel +165e7523fb"></property>
    </properties>
    <testcase classname="simple" name="TestSimpleIngress" time="10.220"></testcase>
    <testcase classname="simple" name="TestSvc2Svc" time="22.820"></testcase>
    <testcase classname="simple" name="TestAuth" time="3.730"></testcase>
  </testsuite>
  <testsuite tests="5" failures="1" time="605.418" name="istio.io/istio/tests/e2e/tests/mixer">
    <properties>
      <property name="go.version" value="devel +165e7523fb"></property>
    </properties>
    <testcase classname="mixer" name="TestGlobalCheckAndReport" time="30.380"></testcase>
    <testcase classname="mixer" name="TestTcpMetrics" time="96.370"></testcase>
    <testcase classname="mixer" name="TestNewMetrics" time="101.190"></testcase>
    <testcase classname="mixer" name="TestDenials" time="12.150"></testcase>
    <testcase classname="mixer" name="TestRateLimit" time="114.310">
        <failure message="Failed" type=""> error message and log </failure>
    </testcase>
  </testsuite>
  <testsuite> ... </testsuite>
</testsuites>
```

Currently, we use [go-junit-report](https://github.com/jstemmer/go-junit-report) to
parse the build log in text into junit report in xml. If there has been a failure
during clean up in between test cases, some test cases would have been skipped. If
there has been clean up error in btween test suites, the `--keep-going` flag when
calling the test target ensures all make targets are executed, and perhaps all test
cases shall pass too, but the clean up error will still fail the job as a whole. The
bottomline is, the exit status / code of the job process is the source of truth and
the junit report is merely a summary of the log.

#### Only job status is reported but no results on test cases

The direct cause for this issue is that the junit report is missing. It can be due
to the fact that artifacts were not successfully uploaded, or that the junit report
was not generated in the first place, or that the generated junit report is not in
valid xml format (which could happen due to special character such as `%`), or that
the junit report is uploaded to the wrong location or the GCS path prefix in the
`config.yaml` used by Testgrid points to the wrong location.
