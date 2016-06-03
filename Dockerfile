FROM ubuntu:wily

RUN apt-get update && \
    apt-get install -y dnsmasq-base git curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN curl -Lo /usr/local/bin/jq https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
  && chmod u+x /usr/local/bin/jq

RUN curl -L https://github.com/gliderlabs/sigil/releases/download/v0.4.0/sigil_0.4.0_Linux_x86_64.tgz \
  | tar -xz -C /usr/local/bin

RUN mkdir -p /etc/mayu /var/lib/mayu /usr/lib/mayu
COPY /bin-dist/mayu /bin-dist/mayuctl /usr/bin/
COPY /bin-dist/tftproot /usr/lib/mayu/tftproot
COPY /bin-dist/static_html /usr/lib/mayu/static_html
COPY template_snippets /usr/lib/mayu/template_snippets
COPY templates /usr/lib/mayu/templates
COPY config.yaml.tmpl /etc/mayu/
COPY ./mayu-entrypoint.sh /

# enable if you want to add a post hook to github to store your cluster config
#RUN ssh-keyscan -H github.com > /etc/ssh/ssh_known_hosts

WORKDIR /usr/lib/mayu

ENTRYPOINT ["/mayu-entrypoint.sh"]
CMD ["-v=12"]
