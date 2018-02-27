FROM alpine

RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*

ADD mason /usr/bin/mason

RUN chmod +x /usr/bin/mason

ENTRYPOINT ["/usr/bin/mason"]


