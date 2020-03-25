#!/usr/bin/env bash

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

# This script will:
#   Create a configmap in a K8s cluster for specifying Istio or Release Builder deps

# e.g.
#  create-deps-cm.sh --context=gke_prow_us-west1-a_prow --branch=master --namespace=test-pods --key=dependencies --interactive master-deps

# NOTE: this only works for istio and release-builder `release-1.5` or newer.

set -o errexit
set -o nounset
set -o pipefail

######################################################################

script_name=$(basename "$0")
branch="master"
context="$(kubectl config current-context)"
namespace="test-pods"
cm_key="dependencies"
local=
interactive=
dry_run=

######################################################################

prompt() {
  local msg="$1" cmd=("${@:2}")

  if [ -z "$interactive" ]; then
    "${cmd[@]}"
    return
  fi

  echo -en "$msg"
  read -r -n1 -p " [y/n] "
  echo

  if [[ $REPLY =~ ^[Yy]$ ]]; then
    "${cmd[@]}"
  else
    {
      echo
      echo "aborting"
      exit 1
    } >&2
  fi
}

# GNU getopt compatibility check
check_getopt() {
  if getopt --test &>/dev/null; then
    {
      echo
      echo "$script_name requires the use of the \"GNU getopt\" command."
      echo
      echo "brew install gnu-getopt"
      echo "brew link --force gnu-getopt"
      exit 1
    } >&2
  fi
}

get_options() {
  check_getopt

  if opt=$(getopt -o b:c:n:k:l:i -l branch:,context:,namespace:,key:,local,interactive,dry-run -n "$script_name" -- "$@"); then
    eval set -- "$opt"
  else
    {
      echo
      echo "unable to parse options"
      exit 1
    } >&2
  fi

  while true; do
    case "$1" in
    -b | --branch)
      branch="$2"
      shift 2
      ;;
    -c | --context)
      context="$2"
      shift 2
      ;;
    -n | --namespace)
      namespace="$2"
      shift 2
      ;;
    -k | --key)
      cm_key="$2"
      shift 2
      ;;
    -l | --local)
      local=1
      shift
      ;;
    -i | --interactive)
      interactive=1
      shift
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    --)
      shift
      get_params "$@"
      break
      ;;
    *)
      {
        echo
        echo "unknown option: $1"
        exit 1
      } >&2
      ;;
    esac
  done
}

get_params() {
  if [ $# -eq 0 ]; then
    {
      echo
      echo "configmap name required"
      exit 1
    } >&2
  fi

  cm_name="$1"
}

template_configmap() {
  template_file="$(mktemp -t template-XXXXXXXXXX)"
  trap 'rm -rf ${template_file:?}' EXIT

  if [ -n "$local" ]; then
    istio_dep="localpath: /home/prow/go/src/istio.io/istio"
  else
    istio_dep="branch: $branch"
  fi

  cat <<EOD >"$template_file"
  istio:
    git: git@github.com:istio-private/istio.git
    $istio_dep
  cni:
    git: git@github.com:istio-private/cni.git
    auto: deps
  api:
    git: https://github.com/istio/api
    auto: modules
  proxy:
    git: git@github.com:istio-private/proxy.git
    auto: deps
  pkg:
    git: https://github.com/istio/pkg
    auto: modules
  client-go:
    git: https://github.com/istio/client-go
    branch: $branch
  gogo-genproto:
    git: https://github.com/istio/gogo-genproto
    branch: $branch
  test-infra:
    git: https://github.com/istio/test-infra
    branch: master
  tools:
    git: https://github.com/istio/tools
    branch: $branch
EOD
}

handle_dry_run() {
  cat <<EOD

Create configmap: "$cm_name" in namespace: "$namespace" of context: "$context" with template:

$(cat "$template_file")

EOD
}

handle_namespace() {
  if ! kubectl --context "$context" get namespaces "$namespace" &>/dev/null; then
    kubectl --context "$context" create namespace "$namespace"
  fi
}

handle_confimap() {
  template_configmap

  if kubectl --context "$context" get configmaps --namespace "$namespace" "$cm_name" &>/dev/null; then
    prompt "Configmap: \"$cm_name\" already exists. Replace" delete_configmap
  fi

  prompt "Create configmap: \"$cm_name\" in namespace: \"$namespace\" of context: \"$context\" with template:\n $(cat "$template_file")\n" create_configmap
}

delete_configmap() {
  kubectl --context "$context" delete configmap --namespace "$namespace" "$cm_name"
}

create_configmap() {
  kubectl --context "$context" create configmap --namespace "$namespace" "$cm_name" --from-file="$cm_key=$template_file"
}

print_usage_example() {
  cat <<EOD

To expose this configuration as an environment variable, add the following env declaration to your spec:

  env:
  - name: DEPENDENCIES
    valueFrom:
      configMapKeyRef:
        name: $cm_name
        key: $cm_key

EOD
}

main() {
  get_options "$@"

  template_configmap

  if [ -n "$dry_run" ]; then
    handle_dry_run
  else
    handle_namespace
    handle_confimap
  fi

  print_usage_example
}

main "$@"
