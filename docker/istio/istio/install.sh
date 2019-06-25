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
  python-requests \
  ruby \
  rubygems \
  ruby-dev \
  software-properties-common \
  unzip \
  wget \
  zip \
  jq

gem install --no-ri --no-rdoc fpm

./install-docker.sh
./install-gcloud.sh
./install-kubectl.sh
./install-golang.sh
./install-helm.sh
./install-protoc.sh
./install-yamllint.sh

apt-get clean
rm -rf /var/lib/apt/lists/*
