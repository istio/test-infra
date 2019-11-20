FROM golang:1.13 as build-env

WORKDIR /go/src/authentikos

COPY . /go/src/authentikos

RUN go get -d -v ./...
RUN go build -o /go/bin/authentikos

FROM gcr.io/distroless/base:22bd467b41e5e656e31db347265fae118db166d9

COPY --from=build-env /go/bin/authentikos /

ENTRYPOINT ["/authentikos"]
