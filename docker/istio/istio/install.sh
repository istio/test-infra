#!/bin/bash

set -eux

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get -qqy --no-install-recommends install \
  apt-transport-https \
  build-essential \
  ca-certificates \
  curl \
  lsb-release \
  python \
  software-properties-common \
  unzip \
  wget

./install-docker.sh
./install-gcloud.sh
./install-golang.sh
./install-gosu.sh
./install-helm.sh
./install-protoc.sh

apt-get clean
rm -rf /var/lib/apt/lists/*
