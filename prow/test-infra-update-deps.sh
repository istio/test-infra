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


#############################################
# Crontab cleanup script triggered by Prow. #
#############################################

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

echo "=== Building Binary that Updates Istio Dependency ==="
bazel build //toolbox/deps_update:update_deps

git config --global user.email "istio.testing@gmail.com"
git config --global user.name "istio-bot"

echo "=== Updating Dependency of Istio ==="
./bazel-bin/toolbox/deps_update/update_deps \
--repo=istio \
--token_file=/etc/github/oauth \
--hub=gcr.io/istio-testing
