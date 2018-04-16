FROM google/cloud-sdk:alpine

RUN apk update \
  && apk upgrade \
  && apk add ca-certificates && rm -rf /var/cache/apk/* \
  && gcloud components install kubectl -q --no-user-output-enabled

ADD mason /usr/bin/mason

RUN chmod +x /usr/bin/mason

ENTRYPOINT ["/usr/bin/mason"]


