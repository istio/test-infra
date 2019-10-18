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

LLVM_VERSION="9.0.0"
LLVM_TARGET="x86_64-linux-gnu-ubuntu-16.04"

LLVM_BASE_URL="http://releases.llvm.org/${LLVM_VERSION}"
LLVM_ARCHIVE="clang+llvm-${LLVM_VERSION}-${LLVM_TARGET}"
LLVM_URL="${LLVM_BASE_URL}/${LLVM_ARCHIVE}.tar.xz"
LLVM_DOWNLOAD_DIRECTORY="/tmp/clang+llvm"
LLVM_DIRECTORY="/usr/lib/llvm-9"

mkdir -p "${LLVM_DOWNLOAD_DIRECTORY}"
wget -q -nc -P "${LLVM_DOWNLOAD_DIRECTORY}" "${LLVM_URL}"
mkdir -p "${LLVM_DIRECTORY}/lib"
tar -C ${LLVM_DIRECTORY} -xJf "${LLVM_DOWNLOAD_DIRECTORY}/${LLVM_ARCHIVE}.tar.xz"

mv ${LLVM_DIRECTORY}/${LLVM_ARCHIVE}/lib/libc++* ${LLVM_DIRECTORY}/lib/
rm -rf ${LLVM_DIRECTORY:?}/${LLVM_ARCHIVE}/

echo "${LLVM_DIRECTORY}/lib" > /etc/ld.so.conf.d/llvm.conf
ldconfig

echo 'libc++ installed.'
