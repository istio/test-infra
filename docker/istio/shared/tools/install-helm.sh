#!/bin/bash

set -eux

HELM_VERSION='v2.10.0'
HELM_BASE_URL="https://storage.googleapis.com/kubernetes-helm"
HELM_ARCHIVE="helm-${HELM_VERSION}-linux-amd64.tar.gz"
HELM_URL="${HELM_BASE_URL}/${HELM_ARCHIVE}"

curl -L ${HELM_URL} | tar xfz - --strip-components=1 -C /usr/local/bin/ linux-amd64/helm

helm -h
