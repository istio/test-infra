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
set -e

print_error() {
    local last_return="$?"

    {
        echo
        echo "${1:-unknown error}"
        echo
    } >&2

    return "${2:-$last_return}"
}

print_error_and_exit() {
    print_error "${1:-unknown error}"
    exit "${2:-1}"
}

RELEASE_NOTES_NONE_LABEL="release-notes-none"

cleanup() {
    rm -rf "${tmp_token:-}"
}

get_opts() {
    if opt="$(getopt -o '' -l token-path:,token:,pr:,org:,repo:,base-sha:,pr-head-sha:,repo-path: -n "$(basename "$0")" -- "$@")"; then
        eval set -- "$opt"
    else
        print_error_and_exit "unable to parse options"
    fi

    while true; do
        case "$1" in
        --token-path)
            token_path="$2"
            token="$(cat "$token_path")"
            shift 2
            ;;
        --token)
            token="$2"
            tmp_token="$(mktemp -t token-XXXXXXXXXX)"
            echo "$token" >"$tmp_token"
            token_path="$tmp_token"
            shift 2
            ;;
        --org)
            REPO_OWNER="$2"
            shift 2
            ;;
        --repo)
            REPO_NAME="$2"
            shift 2
            ;;
        --pr)
            PULL_NUMBER="$2"
            shift 2
            ;;
        --pr-head-sha)
            PULL_PULL_SHA="$2"
            shift 2
            ;;
        --repo-path)
            REPO_PATH="$2"
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

#This script relies on the REPO_OWNER, REPO_NAME, and PULL_NUMBER environment
#variables as defined in https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#job-environment-variables.
validate_opts() {
    if [[ "${JOB_TYPE:-}" == "batch" ]]; then
        echo "Release notes checker not applicable for batch jobs. Skipping"
        exit 0
    fi

    if [ -z "${REPO_OWNER:-}" ]; then
        print_error_and_exit "REPO_OWNER is a required option. It must match the GitHub org."
        exit 1
    fi

    if [ -z "${REPO_NAME:-}" ]; then
        print_error_and_exit "REPO_NAME is a required option. It must match the GitHub repo."
        exit 1
    fi

    if [ -z "${PULL_NUMBER:-}" ]; then
        print_error_and_exit "PULL_NUMBER is a required option. It must match the GitHub pull request number."
        exit 1
    fi

    if [ -z "${PULL_PULL_SHA:-}" ]; then
        echo "PULL_PULL_SHA not specified. This must match the HEAD SHA for the pull request."
        exit 1
    fi

    if [ -z "${REPO_PATH:-}" ]; then
        echo "REPO_PATH not specified. Using current working directory."
        REPO_PATH=$(pwd)
    else
        echo "Using REPO_PATH ${REPO_PATH}"
    fi

}

# Curl the GitHub API to get a list of files for the specified PR. If files are
# found, exit. We might eventually want to validate the data here.
checkForFiles() {
    echo "Checking files from pull request ${REPO_OWNER}/${REPO_NAME}#${PULL_NUMBER}"

    pushd "${REPO_PATH}"

    GEN_RELEASE_NOTES_PATH=../tools/cmd/gen-release-notes
    pushd ${GEN_RELEASE_NOTES_PATH}
    go build
    popd

    set +e
    "${GEN_RELEASE_NOTES_PATH}"/gen-release-notes --pullRequest "${PULL_NUMBER}" --templates "${GEN_RELEASE_NOTES_PATH}"/templates --notes . --validateOnly
    returnCode=$?
    set -e

    # gen-release-notes returns EX_NOINPUT (return code 66) for files not found. Any other error codes or return codes are valid reasons to exit.
    EX_NOINPUT=66
    if [ "${returnCode}" -eq ${EX_NOINPUT} ]; then
        echo "No release notes files found in '/releasenotes/notes/'."
        echo ""
        popd
    else
        #gen-release-notes will either return 0 (no error) or an error.
        exit "${returnCode}"
    fi
}

## Validate the release notes against the schema
function validateNote() {
    out=$(../tools/cmd/schema-validator/schema-validator  --schemaPath ../tools/cmd/gen-release-notes/release_notes_schema.json --documentPath "${1}")
    returnCode=$?
    if [ "${returnCode}" != 0 ]; then
        echo "${out}"
    fi
    return $returnCode
}

function validateNotes() {
    #Build the schema validator
    SCHEMA_VALIDATOR_PATH=../tools/cmd/schema-validator
    pushd ${SCHEMA_VALIDATOR_PATH}
    go build
    popd

    pushd "${REPO_PATH}"
    local errorOccurred=0
    gh pr view "${PULL_NUMBER}" --json files | jq -r '.files[].path' | grep -E '^releasenotes' | \
    {
    set +e
    while read -r line; do
    if ! validateNote "./${line}"; then
        errorOccurred=1
    fi
    popd

    done
    set -e
    if [ "${errorOccurred}" != 0 ]; then
        echo ""
        echo "Failed to validate release notes. Exiting"
        exit "${errorOccurred}"
    fi
    }

}

checkForLabel() {
    # If no release notes files were found in the PR, go back and request the actual
    # PR info (instead of the files) and check the labels.
    echo "Requesting labels from pull request ${REPO_OWNER}/${REPO_NAME}#${PULL_NUMBER}"

    if [ -z "${token}" ]; then
        ghPR=$(curl -L -s --show-error --fail  \
            -H "Accept: application/vnd.github.v3+json" \
            "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/pulls/${PULL_NUMBER}")
    else
        ghPR=$(curl -L -s --show-error --fail \
            --header "Accept: application/vnd.github.v3+json" \
            --header "Authorization: Bearer ${token}" \
            "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/pulls/${PULL_NUMBER}")
    fi

    labels=$(echo "${ghPR}" | jq '.labels | map(.name)')
    owner=$(echo "${ghPR}" | jq -r '.user.login')
    branch=$(echo "${ghPR}" | jq -r '.head.ref')
    repo=$(echo "${ghPR}" | jq -r '.head.repo.name')

    # grep returns a non-zero error code on not found. Reset -e so we don't fail silently.
    set +e
    releaseNotesLabelPresent=$(echo "${labels}" | grep $RELEASE_NOTES_NONE_LABEL)
    set -e

    if [ -z "${releaseNotesLabelPresent}" ]; then
        echo "Missing \"${RELEASE_NOTES_NONE_LABEL}\" label"
        echo
        echo "Missing release notes and missing \"${RELEASE_NOTES_NONE_LABEL}\" label."
        echo "If this pull request contains user facing changes, please create a release note based on the template: https://github.com/istio/istio/blob/master/releasenotes/template.yaml by going here: https://github.com/${owner}/${repo}/new/${branch}/releasenotes/notes"
        echo "Release notes documentation can be found here: https://github.com/istio/istio/tree/master/releasenotes"
        echo "If this pull request has no user facing changes, please add the release-notes-none label to the pull request. Note that the test will have to be manually retriggered after adding the label."
        exit 1
    else
        echo "Found ${RELEASE_NOTES_NONE_LABEL} label. This pull request will not include release notes."
        exit 0
    fi
}

main() {
    trap cleanup EXIT
    get_opts "$@"
    validate_opts

    validateNotes
    checkForFiles
    checkForLabel
    return 1
}

main "$@"
