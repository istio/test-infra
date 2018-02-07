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
UPDATE_EXT_DEP="false"

# List of repo where auto dependency update has been enabled
# excluding istio/istio
case ${GIT_BRANCH} in
  master)
    hour=`date "+%I"` #( 1..12)
    day_of_week=`date "+%u"` #( 1..7)
    repos=( mixerclient proxy )
    case ${hour} in
      02|04|06|08|10|12)
	./bazel-bin/toolbox/deps_update/deps_update \
		--repo="istio" \
		--base_branch=${GIT_BRANCH} \
		--token_file=${TOKEN_PATH} \
		--update_ext_dep="false"
        ;;
      *)
        ;;
    esac
    hour24=`date "+%k"` #( 0..23)
    if [ "${hour24}" -ge 20 ] && [ "${day_of_week}" -eq 2 ]; then
	# external deps (envoyproxy in proxy and depend.update in istio) updated only once a week
        UPDATE_EXT_DEP="true"
    fi
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
    --token_file=${TOKEN_PATH} \
    --update_ext_dep=${UPDATE_EXT_DEP}
done
