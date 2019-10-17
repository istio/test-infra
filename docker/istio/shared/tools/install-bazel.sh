#!/bin/bash

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

set -eux

# Use Bazelisk launcher for Bazel (automatically downloads required version of Bazel).
BAZELISK_VERSION="v1.0"
BAZELISK_BASE_URL="https://github.com/bazelbuild/bazelisk/releases/download"
BAZELISK_BIN="bazelisk-linux-amd64"
BAZELISK_URL="${BAZELISK_BASE_URL}/${BAZELISK_VERSION}/${BAZELISK_BIN}"

wget -q -nc "${BAZELISK_URL}"
chmod +x "${BAZELISK_BIN}"
mv ./${BAZELISK_BIN} /usr/local/bin/bazel

# Preload Bazel versions that we know are needed.
USE_BAZEL_VERSION=0.22.0 /usr/local/bin/bazel version # release-1.1
USE_BAZEL_VERSION=0.26.1 /usr/local/bin/bazel version # release-1.2
USE_BAZEL_VERSION=0.28.1 /usr/local/bin/bazel version # release-1.3
USE_BAZEL_VERSION=1.0.0 /usr/local/bin/bazel version # release-1.4
