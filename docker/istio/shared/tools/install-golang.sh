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

GO_VERSION='1.13'
GO_BASE_URL="https://storage.googleapis.com/golang"
GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="${GO_BASE_URL}/${GO_ARCHIVE}"

KIND_VERSION=v0.5.1

export GOPATH=/opt/go
mkdir -p ${GOPATH}/bin

curl -L "${GO_URL}" | tar -C /usr/local -xzf -

export PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}

go version

go get github.com/github/hub
go get github.com/golang/dep/cmd/dep
go get -u google.golang.org/api/sheets/v4
go get github.com/jstemmer/go-junit-report
GO111MODULE="on" go get sigs.k8s.io/kind@${KIND_VERSION}

rm -rf "${GO_ARCHIVE}"
