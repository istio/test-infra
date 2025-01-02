FROM golang:1.23 as build-env

WORKDIR /go/src/authentikos

COPY . /go/src/authentikos

RUN go build -o /go/bin/authentikos

# safer to use latest than never update...
# hadolint ignore=DL3006
FROM gcr.io/distroless/static

COPY --from=build-env /go/bin/authentikos /

ENTRYPOINT ["/authentikos"]
