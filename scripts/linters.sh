#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${ROOT}/scripts"
UNSET_GOPATH="$(cd "$(dirname "${ROOT})")/../.." && pwd)"
GOPATH=${GOPATH:-${UNSET_GOPATH}}

function install_golangcilint() {
    # if you want to update this version, also change the version number in .golangci.yml
    GOLANGCI_VERSION="v1.16.0"
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b "$GOPATH"/bin "$GOLANGCI_VERSION"
    golangci-lint --version
}

function run_golangcilint() {
  echo 'Running golangcilint ...'
  golangci-lint run --issues-exit-code=0  -v ./...
  echo 'golangcilint OK'
}

function check_licences() {
  echo 'Checking licences ...'
  ${BIN_PATH}/check_license.sh
  echo 'licences OK'
}

function run_buildifier() {
  bazel run //:buildifier -- -showlog -mode=check $(git ls-files| grep -e BUILD -e WORKSPACE | grep -v vendor)
  echo 'buildifier OK'
}

echo "GOPATH is set to ${GOPATH}"

pushd ${ROOT}
install_golangcilint
run_golangcilint
check_licences
run_buildifier
popd
