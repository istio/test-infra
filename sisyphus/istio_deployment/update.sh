#!/bin/bash

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

DEPLOYMENT="istio-toolbox"
SISYPHUS=$(dirname $0)
GIT_HEAD=$(git rev-parse HEAD)
MODIFIED_YAML="${SISYPHUS}/local.sisyphus-deployment.yaml"

bazel run //sisyphus/istio_deployment:sisyphus_image

docker tag bazel/sisyphus/istio_deployment:sisyphus_image \
gcr.io/istio-testing/sisyphus:${GIT_HEAD}

gcloud docker -- push gcr.io/istio-testing/sisyphus:${GIT_HEAD}

sed 's/image: gcr.io\/istio-testing\/sisyphus:.*/image: gcr.io\/istio-testing\/sisyphus:'"$GIT_HEAD"'/g' \
${SISYPHUS}/sisyphus-deployment.yaml > ${MODIFIED_YAML}

cat ${MODIFIED_YAML}

echo "To replace:
1. connect to \"${DEPLOYMENT}\" cluster
2. kubectl replace -f ${MODIFIED_YAML} -n sisyphus"
