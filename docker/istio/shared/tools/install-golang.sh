#!/bin/bash

set -eux

GO_VERSION='1.11.4'
GO_BASE_URL="https://storage.googleapis.com/golang"
GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="${GO_BASE_URL}/${GO_ARCHIVE}"

export GOPATH=/opt/go
mkdir -p ${GOPATH}/bin

curl -L "${GO_URL}" | tar -C /usr/local -xzf -

export PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}

go version

go get github.com/github/hub
go get github.com/golang/dep/cmd/dep
go get github.com/jstemmer/go-junit-report
go get gopkg.in/alecthomas/gometalinter.v2
gometalinter.v2 --install
rm -rf "${GO_ARCHIVE}"
