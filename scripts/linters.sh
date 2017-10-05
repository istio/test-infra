#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${ROOT}/scripts"

bazel ${BAZEL_STARTUP_ARGS} build ${BAZEL_RUN_ARGS} \
  //... $(bazel query 'tests(//...)')

source ${BIN_PATH}/use_bazel_go.sh

cd ${ROOT}
echo 'Running gometalinter ...'
docker run \
  -v $(bazel info output_base):$(bazel info output_base) \
  -v $(pwd):/go/src/istio.io/test-infra \
  -w /go/src/istio.io/test-infra \
  gcr.io/istio-testing/linter:bfcc1d6942136fd86eb6f1a6fb328de8398fbd80 \
  --concurrency=4\
  --enable-gc\
  --vendored-linters\
  --deadline=600s --disable-all\
  --enable=aligncheck\
  --enable=deadcode\
  --enable=errcheck\
  --enable=gas\
  --enable=goconst\
  --enable=gofmt\
  --enable=goimports\
  --enable=golint --min-confidence=0 --exclude=.pb.go --exclude=vendor/ --exclude='should have a package comment'\
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
  ./...
echo 'gometalinter OK'

echo 'Checking licences ...'
${BIN_PATH}/check_license.sh
echo 'licences OK'

echo 'Running buildifier ...'
go get github.com/bazelbuild/buildtools/buildifier
buildifier -showlog -mode=check $(git ls-files| grep -e BUILD -e WORKSPACE | grep -v vendor)
echo 'buildifier OK'
