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
  'cloudtrace.googleapis.com'
)

CREATE_PROJECT=false
BILLING_ACCOUNT=

while getopts :b:p:f:c arg; do
  case ${arg} in
    p) PROJECT_ID="${OPTARG}";;
    c) CREATE_PROJECT=true;;
    b) BILLING_ACCOUNT="${OPTARG}";;
    f) FOLDER="${OPTARG}";;
    *) error_exit "Unrecognized argument -${OPTARG}";;
  esac
done

if [[ ${CREATE_PROJECT} == true ]]; then
  [[ -z "${BILLING_ACCOUNT}" ]] && { echo "use -b to set billing account"; exit 1; }
  [[ -z "${FOLDER}" ]] && { echo "use -f to set folder"; exit 1; }
  gcloud projects create --folder ${FOLDER} "${PROJECT_ID}"
  gcloud alpha billing projects link "${PROJECT_ID}" --billing-account "${BILLING_ACCOUNT}"
fi

for sa in ${USERS[@]}; do
  gcloud projects add-iam-policy-binding ${PROJECT_ID} --member=${sa} --role roles/owner
done

for s in ${SERVICES[@]}; do
  gcloud services enable ${s} --project ${PROJECT_ID}
done

