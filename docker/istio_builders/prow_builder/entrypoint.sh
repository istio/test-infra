#!/bin/bash

## Copyright 2017 Istio Authors
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
##     http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.

#
# entrypoint starts docker daemon and runs bootstrap.py
# In this version, repo tool is in charge of checking out the code,
# bootstrap is just used to start jobs and logs gathering
#

# Prow passes down the following environment variables
# Build settings
# BUILD_NUMBER: Build number
# JOB_NAME: The job name

# Prow supports postsbumits, presubmits and periodic jobs.
# For periodic jobs, nothing else is passed down
# For Postsubmit and batch jobs the following is passed down
# REPO_OWNER: Name of the org
# REPO_NAME: Name of the repo
# PULL_BASE_REF:  The origin branch
# PULL_BASE_SHA: The SHA from the origin branch
# PULL_REFS: Batch mode, comma separated list of PULL_BASE_REF:PULL_BASE_SHA
# For pull request, but we won't use this, as the same information we care about
# is included in PULL_REFS
# PULL_NUMBER: Pull Request number
# PULL_PULL_SHA: SHA of the pull request

if [[ ${UID} -eq 0 ]]; then
  SUDO=
else
  SUDO=sudo
fi

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x
# print all commands that are run

${SUDO} service docker start

# Useful to have in logs for debugging
pwd
env

# start harness (checkout code/run job/upload logs)
bootstrap.py --no-magic-env "$@"
