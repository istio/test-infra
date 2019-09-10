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

set -eux

HELM_VERSION='v2.10.0'
HELM_BASE_URL="https://storage.googleapis.com/kubernetes-helm"
HELM_ARCHIVE="helm-${HELM_VERSION}-linux-amd64.tar.gz"
HELM_URL="${HELM_BASE_URL}/${HELM_ARCHIVE}"

curl -L ${HELM_URL} | tar xfz - --strip-components=1 -C /usr/local/bin/ linux-amd64/helm

helm -h
