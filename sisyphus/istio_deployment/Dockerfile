FROM gcr.io/istio-testing/prowbazel:0.4.8

RUN go get -u istio.io/test-infra/sisyphus/istio_deployment

ENTRYPOINT ["istio_deployment", "--email_sending=false"]
