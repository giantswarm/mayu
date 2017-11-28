FROM alpine:3.6

RUN apk add --no-cache git ca-certificates dnsmasq 

RUN mkdir -p /etc/mayu /var/lib/mayu /usr/lib/mayu
COPY mayu  /mayu
COPY tftproot /usr/lib/mayu/tftproot
COPY static_html /usr/lib/mayu/static_html
COPY template_snippets /usr/lib/mayu/template_snippets
COPY templates /usr/lib/mayu/templates
COPY config.yaml* /etc/mayu/

WORKDIR /usr/lib/mayu

RUN if [ ! -f /etc/mayu/config.yaml ]; then cp /etc/mayu/config.yaml.dist /etc/mayu/config.yaml; fi

ENTRYPOINT ["/mayu"]
CMD ["--cluster-directory=/var/lib/mayu","-v=12"]
