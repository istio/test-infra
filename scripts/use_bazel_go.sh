# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file should be sourced before using go commands
# it ensures that bazel's version of go is used

BAZEL_DIR="$(bazel info execution_root)"

export GOROOT="${BAZEL_DIR}/external/go_sdk"

export PATH=$GOROOT/bin:$PATH

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "*** Calling ${BASH_SOURCE[0]} directly has no effect. It should be sourced."
  echo "Using GOROOT: $GOROOT"
  go version
fi
