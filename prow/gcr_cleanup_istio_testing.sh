#!/bin/bash

# Copyright © 2018 Google Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Thanks to Googler Ahmet Alp Balkan for a basic version
# https://gist.github.com/ahmetb/7ce6d741bd5baa194a3fac6b1fec8bb7


set -eou pipefail

function delete_image() {
  IFS=$'\n\t'
  local C=0
  local R=0
  IMAGE="${1}"
  DATE="${2}"
  echo set -eou pipefail
  echo echo deleting $IMAGE containers older than $DATE
  for digest in $(gcloud container images list-tags ${IMAGE} --limit=999999 --sort-by=TIMESTAMP \
     --filter "timestamp.datetime < '${DATE}'" --format='get(digest)'); do
#    --filter "tags:???????????????????????????????????????? AND timestamp.datetime < '${DATE}'" --format='get(digest)'); do
#--filter "NOT tags:* AND timestamp.datetime < '${DATE}'" --format='get(digest)'); do
    if [ $R -eq 0 ] || [ $R -eq 10 ]; then
      echo ""
      echo -n gcloud container images delete -q --force-delete-tags "${IMAGE}@${digest} "
      R=0
    else
      echo -n                                                       "${IMAGE}@${digest} "
    fi
    let C=C+1
    let R=R+1
  done
  echo ""
  echo "echo # Deleted ${C} images in ${IMAGE}. >&2"
  unset IFS
}


function delete_all_images() {
  local TMP_DIR=$1
  local DEL_DATE=$2
  local REGISTRY="gcr.io/istio-testing"

#  for image_name in galley grafana istio-ca istio-ca-test manager mixer mixer_debug node-agent node-agent-test pilot proxy proxy_debug proxy_init proxyv2 servicegraph servicegraph_debug sidecar_initialzer sidecar_injector test_policybackend testmixer; do

  for image_name in app citadel citadel-test eurekamirror flexvolumedriver fortio fortio.echosrv fortio.fortio fortio.grpcping galley grafana istio-ca istio-ca-test manager mixer mixer_debug node-agent node-agent-test pilot proxy proxy_debug proxy_init proxyv2 servicegraph servicegraph_debug sidecar_initialzer sidecar_injector test_policybackend testmixer; do

    touch     $TMP_DIR/$image_name
    chmod +x  $TMP_DIR/$image_name
    delete_image $REGISTRY/$image_name $DEL_DATE >> $TMP_DIR/$image_name

    echo      echo "\\n\\n" $image_name start
    echo      $TMP_DIR/$image_name
    echo      date
    echo      echo $image_name done
  done
}


TMP_DIR=$(mktemp -d)
[[ ! -z "${TMP_DIR}"  ]] || exit 1

DEL_DATE=$(date "+%C%y-%m-%d" -d "-30 days")

echo     $TMP_DIR/delete_all.sh
touch    $TMP_DIR/delete_all.sh
chmod +x $TMP_DIR/delete_all.sh
delete_all_images $TMP_DIR $DEL_DATE >> $TMP_DIR/delete_all.sh

$TMP_DIR/delete_all.sh


echo    $TMP_DIR
#rm -rf $TMP_DIR
