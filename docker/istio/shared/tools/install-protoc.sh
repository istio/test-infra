#!/bin/bash

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
