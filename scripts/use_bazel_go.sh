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
