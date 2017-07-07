#!/bin/bash

# Applies requisite code formatters to the source tree

set -o errexit
set -o nounset
set -o pipefail

# Prefer gofmt from io_bazel_rules_go_toolchain.
EXTERNAL='bazel-bin/toolbox/format.runfiles'
gofmt="${EXTERNAL}/io_bazel_rules_go_toolchain/bin/gofmt"
goimports="${EXTERNAL}/org_golang_x_tools/cmd/goimports/goimports"
buildifier="${EXTERNAL}/com_github_bazelbuild_buildtools/buildifier/buildifier"

[[ ! -x "${gofmt}" ]] && gofmt=$(which gofmt)
[[ ! -x "${goimports}" ]] && goimports=$(which goimports)
[[ ! -x "${buildifier}" ]] && buildifier=$(which buildifier)

GO_FILES=$(git ls-files | grep  '.*\.go')
UX=$(uname)

#remove blank lines so gofmt / goimports can do their job
for fl in ${GO_FILES}; do
  if [[ ${UX} == "Darwin" ]];then
    sed -i '' -e "/^import[[:space:]]*(/,/)/{ /^\s*$/d;}" $fl
  else
    sed -i -e "/^import[[:space:]]*(/,/)/{ /^\s*$/d;}" $fl
fi
done

${gofmt} -s -w ${GO_FILES}
${goimports} -w -local istio.io ${GO_FILES}
${buildifier} -showlog -mode=fix $(git ls-files | grep -e 'BUILD' -e 'WORKSPACE' -e '.*\.bazel' -e '.*\.bzl')
