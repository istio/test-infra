FROM golang:1.19 as build-env

WORKDIR /go/src/authentikos

COPY . /go/src/authentikos

RUN go build -o /go/bin/authentikos

FROM gcr.io/distroless/base:22bd467b41e5e656e31db347265fae118db166d9

COPY --from=build-env /go/bin/authentikos /

ENTRYPOINT ["/authentikos"]
