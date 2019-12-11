FROM golang:1.13 as build-env

WORKDIR /go/src/istio

COPY . /go/src/istio

RUN go get -d -v ./...
RUN go build -o /go/bin/genjobs /go/src/istio/prow/genjobs

FROM gcr.io/distroless/base:22bd467b41e5e656e31db347265fae118db166d9

COPY --from=build-env /go/bin/genjobs /

ENTRYPOINT ["/genjobs"]
