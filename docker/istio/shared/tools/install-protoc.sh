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

PROTOC_VERSION="3.7.1"
PROTOC_ARCHIVE="protoc-${PROTOC_VERSION}-linux-x86_64.zip"
PROTOC_BASE_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}"
PROTOC_URL="${PROTOC_BASE_URL}/${PROTOC_ARCHIVE}"
PROTOC_DIR="/tmp/protoc_install"

wget -q -nc -O "${PROTOC_ARCHIVE}" "${PROTOC_URL}"
mkdir -p "${PROTOC_DIR}"
unzip "${PROTOC_ARCHIVE}" -d "${PROTOC_DIR}"
cp -fp "${PROTOC_DIR}/bin/protoc" /usr/local/bin/protoc
cp -frp ${PROTOC_DIR}/include/* /usr/local/include/
rm -rf "${PROTOC_ARCHIVE}" "${PROTOC_DIR}"

export PATH=${PATH}:/usr/local/bin
