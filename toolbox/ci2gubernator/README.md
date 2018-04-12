## Introduction

Gubernator and testgrid are both frontends that display test results stored in GCS.

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

At the start of a build, execute the following command to create and upload `started.json` on GCS.

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

Caveat: this command exits with the same exit code as supplied by --exit_code.

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
