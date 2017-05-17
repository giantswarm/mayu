FROM alpine:3.4

RUN apk add --update ca-certificates \
    && rm -rf /var/cache/apk/*

ADD ./mayu-infopusher /mayu-infopusher

ENTRYPOINT ["/mayu-infopusher"]

CMD [ "--help" ]
