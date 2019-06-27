#!/bin/bash

set -eux

BAZEL_VERSION='0.27.0'
BAZEL_BASE_URL='https://github.com/bazelbuild/bazel/releases/download'
BAZEL_SH="bazel-${BAZEL_VERSION}-installer-linux-x86_64.sh"
BAZEL_URL="${BAZEL_BASE_URL}/${BAZEL_VERSION}/${BAZEL_SH}"

wget -q -nc "${BAZEL_URL}"
chmod +x "${BAZEL_SH}"
./${BAZEL_SH}
rm -rf "${BAZEL_SH}"
