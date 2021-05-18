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

get_opts() {
    if opt="$(getopt -o '' -l document-path:,schema-path:,repo-path: -n "$(basename "$0")" -- "$@")"; then
        eval set -- "$opt"
    else
        print_error_and_exit "unable to parse options"
    fi

    while true; do
        case "$1" in
        --document-path)
            DOCUMENT_PATH="$2"
            shift 2
            ;;
        --schema-path)
            SCHEMA_PATH="$2"
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

validate_opts() {
    if [ -z "${DOCUMENT_PATH:-}" ]; then
        print_error_and_exit "DOCUMENT_PATH is a required option. It must be the path to the document to validate."
        exit 1
    fi

    if [ -z "${SCHEMA_PATH:-}" ]; then
        print_error_and_exit "SCHEMA_PATH is a required option. It must be the path to the schema to validate against."
        exit 1
    fi
}

# Curl the GitHub API to get a list of files for the specified PR. If files are
# found, exit. We might eventually want to validate the data here.
validate_schema() {
    echo "Validating ${DOCUMENT_PATH} against ${SCHEMA_PATH}"

    pushd "${REPO_PATH}"

    VALIDATE_SCHEMA_PATH=../tools/cmd/schema-validator
    pushd ${VALIDATE_SCHEMA_PATH}
    go build
    popd

    set +e
    "${VALIDATE_SCHEMA_PATH}"/schema-validator --documentPath "${DOCUMENT_PATH}" --schemaPath "${SCHEMA_PATH}"
    returnCode=$?
    set -e

    exit "${returnCode}"
}

main() {
    get_opts "$@"
    validate_opts

    validate_schema
    return 1
}

main "$@"
