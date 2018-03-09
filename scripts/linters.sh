#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${ROOT}/scripts"

bazel ${BAZEL_STARTUP_ARGS} build ${BAZEL_RUN_ARGS} \
  //... $(bazel query 'tests(//...)')

source ${BIN_PATH}/use_bazel_go.sh

echo 'Installing gometalinter ...'
go get -u gopkg.in/alecthomas/gometalinter.v2

gometalinter.v2 --install

cd ${ROOT}
echo 'Running gometalinter ...'
gometalinter.v2 ./... \
  --concurrency=4\
  --deadline=600s --disable-all\
  --enable=deadcode\
  --enable=errcheck\
  --enable-gc\
  --enable=goconst\
  --enable=gofmt\
  --enable=goimports\
  --enable=golint --min-confidence=0\
  --enable=ineffassign\
  --enable=interfacer\
  --enable=lll --line-length=160\
  --enable=megacheck\
  --enable=misspell\
  --enable=structcheck\
  --enable=unconvert\
  --enable=varcheck\
  --enable=vet\
  --enable=vetshadow\
  --exclude='should have a package comment'\
  --exclude=.pb.go\
  --vendor\
  --vendored-linters\
  ./...
echo 'gometalinter OK'

echo 'Checking licences ...'
${BIN_PATH}/check_license.sh
echo 'licences OK'

echo 'Running buildifier ...'
go get github.com/bazelbuild/buildtools/buildifier
buildifier -showlog -mode=check $(git ls-files| grep -e BUILD -e WORKSPACE | grep -v vendor)
echo 'buildifier OK'
