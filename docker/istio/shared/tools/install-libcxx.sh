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

LLVM_VERSION=9
LLVM_REPO='xenial'

echo "deb http://apt.llvm.org/${LLVM_REPO}/ llvm-toolchain-${LLVM_REPO}-${LLVM_VERSION} main" \
    | tee /etc/apt/sources.list.d/llvm.list
echo 'Adding repo for libc++'
curl https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add -
echo 'Installing libc++'
apt-get update
apt-get -qqy install "libc++-${LLVM_VERSION}-dev" "libc++abi-${LLVM_VERSION}-dev"
