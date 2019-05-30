#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${ROOT}/scripts"
UNSET_GOPATH="$(cd "$(dirname "${ROOT})")/../.." && pwd)"
GOPATH=${GOPATH:-${UNSET_GOPATH}}

function install_gometalinter() {
  echo 'Installing gometalinter ...'
  curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.16.0
}

function run_gometalinter() {
  echo 'Running gometalinter ...'
  "${GOPATH}/bin/golangci-lint" run ./... \
    --concurrency=4\
    --deadline=600s --disable-all\
    --enable=deadcode\
    --enable=errcheck\
    --enable=goconst\
    --enable=gofmt\
    --enable=goimports\
    --enable=golint min-confidence=0\
    --enable=ineffassign\
    --enable=interfacer\
    --enable=lll line-length=160\
    --enable=megacheck\
    --enable=misspell\
    --enable=structcheck\
    --enable=unconvert\
    --enable=varcheck\
    --enable=vet\
    --enable=vetshadow\
    --exclude='should have a package comment'\
    --exclude=.pb.go\
    ./...
  echo 'gometalinter OK'
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
install_gometalinter
run_gometalinter
check_licences
run_buildifier
popd
