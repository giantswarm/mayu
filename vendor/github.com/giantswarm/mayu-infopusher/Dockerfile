FROM frolvlad/alpine-glibc

RUN apk add --no-cache ca-certificates \
    && mkdir /embedded

ADD ./mayu-infopusher /mayu-infopusher
ADD ./embedded/ipmitool /embedded/

ENTRYPOINT ["/mayu-infopusher"]
CMD [ "--help" ]
