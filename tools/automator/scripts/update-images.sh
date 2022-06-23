#!/usr/bin/env bash
# shellcheck disable=SC2016

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

ROOT="$(cd -P "$(dirname -- "$0")" && pwd -P)"

# shellcheck disable=SC1090,SC1091
source "$ROOT/../utils.sh"

# Defaults
image='gcr.io/istio-testing/build-tools.*'
tag='$AUTOMATOR_SRC_BRANCH-[0-9a-f]{40}'
paths='$AUTOMATOR_REPO_DIR/prow/cluster/jobs/**/*.yaml,$AUTOMATOR_REPO_DIR/prow/config/jobs/**/*.yaml'
key="image"
var="IMAGE_VERSION"
source='$AUTOMATOR_ROOT_DIR/files/Makefile'

get_opts() {
  if opt="$(getopt -o '' -l pre:,post:,image:,tag:,paths:,key:,var:,source: -n "$(basename "$0")" -- "$@")"; then
    eval set -- "$opt"
  else
    print_error_and_exit "unable to parse options"
  fi

  while true; do
    case "$1" in
    --pre)
      pre="$2"
      shift 2
      ;;
    --post)
      post="$2"
      shift 2
      ;;
    --image)
      image="$2"
      shift 2
      ;;
    --tag)
      tag="$2"
      shift 2
      ;;
    --paths)
      paths=$2
      shift 2
      ;;
    --key)
      key="$2"
      shift 2
      ;;
    --var)
      var="$2"
      shift 2
      ;;
    --source)
      source="$2"
      shift 2
      ;;
    --)
      shift
      break
      ;;
    *)
      print_error_and_exit "unknown option: $1"
      ;;
    esac
  done
}

resolve() {
  image="$(evaluate_tmpl "$image")"
  tag="$(evaluate_tmpl "$tag")"
  paths="$(evaluate_tmpl "$paths")"
  source="$(evaluate_tmpl "$source")"

  resolved_tag="$(grep -Pom1 "$var.*?\K$tag" "$source")"
  resolved_paths="$(split_on_commas "$paths")"
}

work() {
  # Update build-tools and build-tools-centos images
  # shellcheck disable=SC2086
  sed -Ei "s|($key:\s+$image:)$tag|\1$resolved_tag|g" $resolved_paths
}

main() {
  get_opts "$@"
  resolve
  ${pre:-}
  work
  ${post:-}
}

main "$@"
