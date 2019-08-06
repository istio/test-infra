#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# See https://misc.flogisoft.com/bash/tip_colors_and_formatting

color-image() { # Bold magenta
  echo -e "\x1B[1;35m${*}\x1B[0m"
}

color-version() { # Bold blue
  echo -e "\x1B[1;34m${*}\x1B[0m"
}

color-target() { # Bold cyan
  echo -e "\x1B[1;33m${*}\x1B[0m"
}

if command -v gsed &>/dev/null; then
  SED="gsed"
else
  SED="sed"
fi

if ! (${SED} --version 2>&1 | grep -q GNU); then
  # darwin is great (not)
  echo "!!! GNU sed is required.  If on OS X, use 'brew install gnu-sed'." >&2
  exit 1
fi

TAC=tac

if command -v gtac &>/dev/null; then
  TAC=gtac
fi

if ! command -v "${TAC}" &>/dev/null; then
  echo "tac (reverse cat) required. If on OS X then 'brew install coreutils'." >&2
  exit 1
fi

cd "$(git rev-parse --show-toplevel)"

usage() {
  echo "Usage: $(basename "$0") [--list || --latest || vYYYYMMDD-deadbeef] [image subset...]" >&2
  exit 1
}

if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  echo "Detected GOOGLE_APPLICATION_CREDENTIALS, activating..." >&2
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}"
  gcloud auth configure-docker
fi

cmd=
if [[ $# != 0 ]]; then
  cmd="$1"
  shift
fi

# List the $1 most recently pushed prow versions
list-options() {
  count="$1"
  gcloud container images list-tags gcr.io/k8s-prow/plank --limit="${count}" --format='value(tags)' \
      | grep -o -E 'v[^,]+' | "${TAC}"
}

upstream-version() {
 local branch="https://raw.githubusercontent.com/kubernetes/test-infra/master"
 local file="prow/cluster/deck_deployment.yaml"

 curl "$branch/$file" | grep image: | grep -o -E 'v[-0-9a-f]+'
}

# Print 10 most recent prow versions, ask user to select one, which becomes new_version
list() {
  echo "Listing recent versions..." >&2
  echo "Recent versions of prow:" >&2
  mapfile -t options < <(list-options 10)
  if [[ -z "${options[*]}" ]]; then
    echo "No versions found" >&2
    exit 1
  fi
  def_opt=$(upstream-version)
  new_version=
  for o in "${options[@]}"; do
    if [[ "$o" == "$def_opt" ]]; then
      echo -e "  $(color-image "$o"   '*' prow.k8s.io)"
    else
      echo -e "  $(color-version "${o}")"
    fi
  done
  read -rp "Select version [$(color-image "${def_opt}")]: " new_version
  if [[ -z "${new_version:-}" ]]; then
    new_version="${def_opt}"
  else
    found=
    for o in "${options[@]}"; do
      if [[ "${o}" == "${new_version}" ]]; then
        found=yes
        break
      fi
    done
    if [[ -z "${found}" ]]; then
      echo "Invalid version: ${new_version}" >&2
      exit 1
    fi
  fi
}

if [[ -z "${cmd}" || "${cmd}" == "--list" ]]; then
  list
elif [[ "${cmd}" =~ v[0-9]{8}-[a-f0-9]{6,9} ]]; then
  new_version="${cmd}"
elif [[ "${cmd}" == "--latest" ]]; then
  new_version="$(list-options 1)"
elif [[ "${cmd}" == "--auto" ]]; then
  new_version="$(upstream-version)"
else
  usage
fi

${SED} -i "s/\(.*:\)v[a-f0-9-]\+/\1${new_version}/I" prow/cluster/*.yaml prow/config.yaml
