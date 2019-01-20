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

set -o nounset
set -o errexit
set -o pipefail
set -o xtrace

cd "$(bazel info workspace)"
trap 'echo "FAILED" >&2' ERR

mkdir -p ./vendor
#find vendor -type f -name BUILD -o -name BUILD.bazel -exec rm -rf {} \;
touch ./vendor/BUILD.bazel
bazel run //:gazelle
