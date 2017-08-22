#!/bin/bash
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_PATH="${ROOT}/scripts"

bazel ${BAZEL_STARTUP_ARGS} build ${BAZEL_RUN_ARGS} \
  //... $(bazel query 'tests(//...)')

source ${BIN_PATH}/use_bazel_go.sh

cd ${ROOT}

PARENT_BRANCH=''

while getopts :s: arg; do
  case ${arg} in
    s) LAST_GOOD_GITSHA="${OPTARG}";;
    *) error_exit "Unrecognized argument ${OPTARG}";;
  esac
done

prep_linters() {
  if ! which codecoroner > /dev/null; then
    echo "Preparing linters"
    go get -u github.com/alecthomas/gometalinter
    go get -u github.com/bazelbuild/buildifier/buildifier
    go get -u github.com/3rf/codecoroner
    go get -u honnef.co/go/tools/cmd/megacheck
    gometalinter --install --vendored-linters >/dev/null
  fi
}

go_metalinter() {
  if [[ ! -z ${GITHUB_PR_TARGET_BRANCH} ]]; then
    local parent_branch='parent'
    git fetch origin "refs/heads/${GITHUB_PR_TARGET_BRANCH}:${parent_branch}"
    LAST_GOOD_GITSHA="$(git log ${parent_branch}.. --pretty="%H"|tail -1)"
    [[ ! -z ${LAST_GOOD_GITSHA} ]] && LAST_GOOD_GITSHA="${LAST_GOOD_GITSHA}^"
  fi

  # default: lint everything. This runs on the main build
  PKGS=($(bazel query 'kind("go_library", //... except deps(//vendor/...))' \
    | cut -d ':' -f 1 | sort | uniq | sed -e 's,//,./,g')) \
    || PKGS=()

  # convert LAST_GOOD_GITSHA to list of packages.
  if [[ ! -z ${LAST_GOOD_GITSHA} ]];then
    echo "Using ${LAST_GOOD_GITSHA} to compare files to."
    CHANGED_DIRS=($(for fn in $(git diff --name-only ${LAST_GOOD_GITSHA}); do fd="./${fn%/*}"; [ -d ${fd} ] && echo $fd; done | sort | uniq))
    NEW_PKGS=()
    for d in ${CHANGED_DIRS[@]}; do
      for p in ${PKGS}; do
        if [[ ${d} =~ ${p} ]]; then
          NEW_PKGS+=(${d})
        fi
      done
    done
    PKGS=${NEW_PKGS[@]}
  else
    echo 'Running linters on all files.'
  fi

  [[ -z ${PKGS} ]] && return
  # updated to avoid WARNING: staticcheck, gosimple, and unused are all set, using megacheck instead
  gometalinter\
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
    --enable=golint --min-confidence=0 --exclude=.pb.go --exclude=pkg/config/proto/combined.go --exclude="should have a package comment"\
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
    ${PKGS}

  # TODO: These generate warnings which we should fix, and then should enable the linters
  # --enable=dupl\
  # --enable=gocyclo\
  #
  # This doesn't work with our source tree for some reason, it can't find vendored imports
  # --enable=gotype\
}



run_linters() {
  echo Running linters
  buildifier -showlog -mode=check $(find . -name BUILD -type f)
  go_metalinter
  ${BIN_PATH}/check_license.sh
  buildifier -showlog -mode=check $(find . -type f \( -name 'BUILD' -or \
    -name 'WORKSPACE' -or \
    -wholename '.*\.bazel' -or \
    -wholename '.*bzl' \) -print )

  # TODO: Enable this once more of mixer is connected and we don't
  # have dead code on purpose
  # codecoroner funcs ./...
  # codecoroner idents ./...
}

prep_linters

run_linters

echo Done running linters
