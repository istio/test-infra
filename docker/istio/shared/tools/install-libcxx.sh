#!/bin/bash

set -eux

LLVM_VERSION=7
LLVM_REPO='xenial'

echo "deb http://apt.llvm.org/${LLVM_REPO}/ llvm-toolchain-${LLVM_REPO}-${LLVM_VERSION} main" \
    | tee /etc/apt/sources.list.d/llvm.list
echo 'Adding repo for libc++'
curl https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add -
echo 'Installing libc++'
apt-get update
apt-get -qqy install "libc++-${LLVM_VERSION}-dev" "libc++abi-${LLVM_VERSION}-dev"
