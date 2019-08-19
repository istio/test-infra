#!/bin/bash

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
