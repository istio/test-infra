FROM google/cloud-sdk:alpine

# hadolint ignore=DL3017,DL3018,DL3019
RUN apk update \
  && apk upgrade \
  && apk add ca-certificates && rm -rf /var/cache/apk/* \
  && gcloud components install kubectl -q --no-user-output-enabled

COPY mason /usr/bin/mason

RUN chmod +x /usr/bin/mason

ENTRYPOINT ["/usr/bin/mason"]


