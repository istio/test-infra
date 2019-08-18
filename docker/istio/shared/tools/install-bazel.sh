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
