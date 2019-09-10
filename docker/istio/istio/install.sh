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
  jq \
  iptables

gem install --no-ri --no-rdoc fpm

./install-docker.sh
./install-gcloud.sh
./install-kubectl.sh
./install-golang.sh
./install-helm.sh
./install-protoc.sh
./install-yamllint.sh
./install-libcxx.sh

apt-get clean
rm -rf /var/lib/apt/lists/*
