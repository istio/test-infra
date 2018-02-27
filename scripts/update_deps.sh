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
set -e
# Check unset variables
set -u
# Print commands
set -x

WORKSPACE=$(bazel info workspace)

bazel run //:gazelle

go get -u github.com/golang/dep/cmd/dep

dep ensure -v
dep prune -v

find ${WORKSPACE}/vendor -type f -name BUILD -exec rm -rf {} \;
find ${WORKSPACE}/vendor -type f -name BUILD.bazel -exec rm -rf {} \;

${WORKSPACE}/scripts/deps_to_bazel.py

