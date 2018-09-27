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
  python-yaml \
  software-properties-common \
  unzip \
  wget

./install-bazel.sh
./install-docker.sh
./install-gcloud.sh
./install-golang.sh
./install-gosu.sh

apt-get clean
rm -rf /var/lib/apt/lists/*
