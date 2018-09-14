#!/bin/bash

GO_VERSION='1.10.4'
GO_BASE_URL="https://storage.googleapis.com/golang"
GO_ARCHIVE="go${GO_VERSION}.linux-amd64.tar.gz"
GO_URL="${GO_BASE_URL}/${GO_ARCHIVE}"

export GOPATH=/opt/go
mkdir -p ${GOPATH}/bin

export PATH=${GOPATH}/bin:usr/local/go/bin:${PATH}

wget -q -nc "${GO_URL}"
tar -C /usr/local -xzf "${GO_ARCHIVE}"

go get github.com/github/hub
go get github.com/golang/dep/cmd/dep
go get github.com/jstemmer/go-junit-report
go get gopkg.in/alecthomas/gometalinter.v2
gometalinter.v2 --install
rm -rf "${GO_ARCHIVE}"
