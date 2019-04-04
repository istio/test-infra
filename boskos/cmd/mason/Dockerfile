FROM google/cloud-sdk:alpine

RUN apk add --no-cache git make musl-dev go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

RUN apk update \
  && apk upgrade \
  && apk add ca-certificates && rm -rf /var/cache/apk/* \
  && gcloud components install kubectl -q --no-user-output-enabled \
  && go get istio.io/test-infra/boskos/cmd/mason_client

ADD mason /usr/bin/mason

RUN chmod +x /usr/bin/mason

ENTRYPOINT ["/usr/bin/mason"]


