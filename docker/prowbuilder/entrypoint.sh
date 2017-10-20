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

# The prow job should specify the following ENV
# MANIFEST_URL: Repo Manifest URL
# REPO_PATH: Path to patch, this should be where the acual repo is checked out

function set_git() {
  git config --global user.name 'Testing User'
  git config --global user.email 'istio.testing@gmail.com'
}

function apply_patches() {
  # In order to be compatible with batch mode we need to apply all the refs one at
  # a time, and fail in case of conflict.
  pushd ${REPO_PATH}
  # repo should be setting a remote
  local remote=$(git remote | head)
  [[ -n ${remote} ]] || { echo 'Could not find a remote'; exit 1; }
  IFS=',' read -r -a REFS <<< "${PULL_REFS}"
  for r in ${REFS[@]}; do
    IFS=':' read pr sha  <<< "${r}"
    if [[ ${pr} =~ "${PULL_BASE_REF}" ]]; then
      # The first element of the list is the origin branch where the PR would be
      # merged to, or where the postsubmit should start.
      git fetch "${remote}" "${sha}:test" -f
      git checkout test
    else
      echo "Patching ${pr} with SHA: ${sha}"
      # Fetching PR refspec
      git fetch "${remote}" "+refs/pull/${pr}/head:refs/remotes/origin/pr/${pr}" \
        || { echo "Could not fetch refspec for pr ${pr}"; exit 1; }
      # Merging PR to local checkout
      git merge --no-ff -m "Merge PR ${pr}" "${sha}" \
        || { echo "Could not apply PR ${pr}"; exit 1; }
      echo "Successfully patched ${pr}"
    fi
  done
  echo "Patches applied successfully"
  popd
}

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

# Creating directory for bazel caching
mkdir -p ${REPO_NAME}
cd ${REPO_NAME}

# Checking out code with repo
repo init -u ${MANIFEST_URL} -b ${PULL_BASE_REF}
repo sync -c

# For postsubmit and periodic we can just have repo do all the work
# However for presubmit we need to apply the pull request.
if [[ -n ${PULL_NUMBER} ]]; then
  set_git
  apply_patches
fi

# start harness (checkout code/run job/upload logs)
${HOME}/bootstrap.py \
    --job=${JOB_NAME} \
    --bare \
    --service-account=${GOOGLE_APPLICATION_CREDENTIALS} \
    --upload="gs://istio-prow/" \
    --no-magic-env \
    --jobs-dir="${REPO_PATH}/prow" \
    "$@"
