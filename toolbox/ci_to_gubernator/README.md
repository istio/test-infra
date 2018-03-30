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

At the start of a build

```bash
$ ci_to_gubernator --job_starts \
	--sha=<PULL_REFS> \
	--org=istio \
	--repo=istio \
	--pr_number=3576
```
