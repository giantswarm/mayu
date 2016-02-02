FROM ubuntu:wily

ENV DNSMASQ_VERSION 2.75-1
ENV GIT_VERSION 1:2.5.0-1

RUN apt-get update && \
    apt-get install -y \
    dnsmasq=${DNSMASQ_VERSION} \
    git=${GIT_VERSION}

RUN mkdir -p /opt/mayu /opt/mayu/config
WORKDIR /opt/mayu

COPY ./bin-dist /opt/mayu

ENTRYPOINT ["/opt/mayu/mayu", "-config=/opt/mayu/config/config.yaml"]
CMD ["-v=12"]
