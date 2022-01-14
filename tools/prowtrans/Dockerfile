FROM golang:1.13 as build-env

WORKDIR /go/src/istio

COPY . /go/src/istio

RUN go get -d -v ./...
RUN go build -o /go/bin/prowtrans /go/src/istio/tools/prowtrans/cmd/prowtrans

FROM gcr.io/distroless/base:22bd467b41e5e656e31db347265fae118db166d9

COPY --from=build-env /go/bin/prowtrans /

ENTRYPOINT ["/prowtrans"]
