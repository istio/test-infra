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

bazel run //toolbox/sisyphus:sisyphus_image

docker tag bazel/toolbox/sisyphus:sisyphus_image \
gcr.io/istio-testing/sisyphus:${GIT_HEAD}

gcloud docker -- push gcr.io/istio-testing/sisyphus:${GIT_HEAD}

sed 's/image: gcr.io\/istio-testing\/sisyphus:.*/image: gcr.io\/istio-testing\/sisyphus:'"$GIT_HEAD"'/g' \
${SISYPHUS}/alert-deployment.yaml > ${SISYPHUS}/local.alert-deployment.yaml

cat ${SISYPHUS}/local.alert-deployment.yaml

echo "To replace:
1. connect to \"${DEPLOYMENT}\" cluster
2. kubectl replace -f ${SISYPHUS}/local.alert-deployment.yaml -n sisyphus"
