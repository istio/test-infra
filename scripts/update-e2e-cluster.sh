#!/bin/bash

# Copyright 2017 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#!/bin/bash

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x
#

PROJECT_NAME='istio-testing'
ZONE='us-east4-c'
MACHINE_TYPE='n1-standard-4'
NUM_NODES='10'

PROW_CLUSTER='prow'
PROW_ZONE='us-west1-a'
PROW_PROJECT='istio-testing'
PROW_TEST_NS='test-pods'
CONFIG_BACKUPED=false

cleanup () {
  if [ "${CONFIG_BACKUPED}" = true ]; then
    mv "${CONFIG_BACKUP}" "${CONFIG_FILE}"
  fi
  if [ ! -z "${TEMP_DIR}" ] && [ -d "${TEMP_DIR}" ]; then
    rm -r ${TEMP_DIR}
  fi
}
trap cleanup EXIT

while getopts :r:z:n: arg; do
  case ${arg} in
    r) REPO=${OPTARG};;
    z) ZONE=${OPTARG};;
    n) NUM_NODES=${OPTARG};;
    *) error_exit "Unrecognized argument -${OPTARG}";;
  esac
done

if [ "${REPO}" != 'auth' ] && [ "${REPO}" != 'broker' ] && [ "${REPO}" != 'istio' ] && [ "${REPO}" != 'mixer' ] && [ "${REPO}" != 'pilot' ]; then
  echo 'Must specific a repo and it must be auth/brokeristio/pilot/mixer'
  exit 1
fi

# Generate cluster version and name
CLUSTER_VERSION=$(gcloud container get-server-config --project="${PROJECT_NAME}" --zone="${ZONE}" --format='value(defaultClusterVersion)')
echo "Default cluster version: ${CLUSTER_VERSION}"

# Backup original config on your machine
CONFIG_FILE="${HOME}/.kube/config"
CONFIG_BACKUP="${HOME}/.kube/config.backup"
mv ${CONFIG_FILE} ${CONFIG_BACKUP}
CONFIG_BACKUPED=true

# Try to create a rotation cluster, named $REPO-e2e-rbac-rotation-<suffix>, suffix can be 1 or 2
gcloud config set container/use_client_certificate True
for i in {1..2}
do
  CLUSTER_NAME="${REPO}-e2e-rbac-rotation-${i}"
  result=$(gcloud container clusters create ${CLUSTER_NAME} --zone ${ZONE} --project ${PROJECT_NAME} --cluster-version ${CLUSTER_VERSION} \
  --machine-type ${MACHINE_TYPE} --num-nodes ${NUM_NODES} --no-enable-legacy-authorization --enable-kubernetes-alpha --quiet \
  ||  echo 'Failed')
  [[ ${result} == 'Failed' ]] || break
  if [ ${i} -eq 2 ]; then
    echo "Cannot create a rotation cluster for ${REPO}"; exit 1
  fi
done

# Keep new config into temp dir and put original config back
TEMP_DIR="$(mktemp -d)"
mkdir ${TEMP_DIR}
mv "${CONFIG_FILE}" "${TEMP_DIR}/config"
mv "${CONFIG_BACKUP}" "${CONFIG_FILE}"
CONFIG_BACKUPED=false

# Switch to prow cluster
gcloud container clusters get-credentials ${PROW_CLUSTER} --zone ${PROW_ZONE} --project ${PROW_PROJECT}

# Update kubeconfig
SECRET_NAME="${REPO}-e2e-rbac-kubeconfig"
kubectl delete secret ${SECRET_NAME} -n ${PROW_TEST_NS}
sleep 5
kubectl -n ${PROW_TEST_NS} create secret generic ${SECRET_NAME} --from-file=${TEMP_DIR}
kubectl get secret -n ${PROW_TEST_NS}
