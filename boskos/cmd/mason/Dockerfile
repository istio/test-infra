FROM google/cloud-sdk:alpine

RUN apk update \
  && apk upgrade \
  && apk add ca-certificates && rm -rf /var/cache/apk/* \
  && gcloud components install kubectl -q --no-user-output-enabled

RUN gcloud components install kubectl -q

ADD mason /usr/bin/mason

RUN chmod +x /usr/bin/mason

ENTRYPOINT ["/usr/bin/mason"]


