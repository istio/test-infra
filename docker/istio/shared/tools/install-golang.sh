#!/bin/bash

set -eux

GO_VERSION='1.12.5'
GO_BASE_URL="https://storage.googleapis.com/golang"
GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="${GO_BASE_URL}/${GO_ARCHIVE}"

# Remove kind
# KIND_RELEASE="https://github.com/kubernetes-sigs/kind/releases/download/v0.3.0/kind-linux-amd64"

export GOPATH=/opt/go
mkdir -p ${GOPATH}/bin

curl -L "${GO_URL}" | tar -C /usr/local -xzf -

export PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}

go version

go get github.com/github/hub
go get github.com/golang/dep/cmd/dep
go get github.com/jstemmer/go-junit-report
# curl -L "${KIND_RELEASE}" > "${GOPATH}/bin/kind"
# chmod u+x "${GOPATH}/bin/kind"

rm -rf "${GO_ARCHIVE}"
