#!/bin/bash

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
# Usage: ./build-kind-image.sh tag1,tag2,tag3

if [ -z "$1" ]; then
  echo "Usage: $0 tag1,tag2,tag3"
  exit 1
fi

IFS=',' read -ra TAGS <<< "$1"

for TAG in "${TAGS[@]}"; do
  SRC_IMAGE="kindest/node:$TAG"
  DEST_IMAGE="gcr.io/istio-testing/kind-node:$TAG"

  echo "Pulling $SRC_IMAGE..."
  docker pull "$SRC_IMAGE" || { echo "Failed to pull $SRC_IMAGE"; exit 1; }

  echo "Tagging $SRC_IMAGE as $DEST_IMAGE..."
  docker tag "$SRC_IMAGE" "$DEST_IMAGE" || { echo "Failed to tag $SRC_IMAGE"; exit 1; }

  echo "Pushing $DEST_IMAGE..."
  docker push "$DEST_IMAGE" || { echo "Failed to push $DEST_IMAGE"; exit 1; }

  echo "Done with tag: $TAG"
done
