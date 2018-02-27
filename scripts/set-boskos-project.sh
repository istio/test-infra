#!/bin/bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIR="${ROOT}/scripts"

. ${DIR}/all-utilities || { echo "Cannot load Bash utilities" ; exit 1 ; }

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

USERS=(
  'serviceAccount:boskos@istio-testing.iam.gserviceaccount.com'
  'serviceAccount:istio-prow-test-job@istio-testing.iam.gserviceaccount.com'
  'group:mdb.istio-testing@google.com'
)

SERVICES=(
  'compute.googleapis.com'
  'container.googleapis.com'
)

while getopts :p:c arg; do
  case ${arg} in
    p) PROJECT_ID="${OPTARG}";;
    *) error_exit "Unrecognized argument -${OPTARG}";;
  esac
done

for sa in ${USERS[@]}; do
  gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=${sa} --role roles/owner
done

for s in ${SERVICES[@]}; do
  gcloud services enable ${s} --project ${PROJECT_ID}
done

