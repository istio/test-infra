#!/bin/bash

# Copyright 2017 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


#######################################################
# Crontab dependency update script triggered by Prow. #
#######################################################

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

echo "=== Building Binary that Updates Istio Dependency ==="
bazel build //toolbox/deps_update:deps_update

git config --global user.email "istio.testing@gmail.com"
git config --global user.name "istio-bot"

TOKEN_PATH="/etc/github/oauth"
# List of repo where auto dependency update has been enabled
# excluding istio/istio
case ${GIT_BRANCH} in
  master)
    repos=( mixerclient proxy )
    # TODO need to enable for istio/istio
    ;;
  release-0.2)
    repos=( old_mixer_repo mixerclient old_pilot_repo proxy )
    echo "=== Updating Dependency of Istio ==="
    ./bazel-bin/toolbox/deps_update/deps_update \
      --repo=istio \
      --token_file=${TOKEN_PATH} \
      --base_branch=${GIT_BRANCH} \
      --hub=gcr.io/istio-testing
    ;;
  *)
    echo error GIT_BRANCH set incorrectly; exit 1
    ;;
esac

for r in "${repos[@]}"; do
  echo "=== Updating Dependency of ${r} ==="
  ./bazel-bin/toolbox/deps_update/deps_update \
    --repo=${r} \
    --base_branch=${GIT_BRANCH} \
    --token_file=${TOKEN_PATH}
done
