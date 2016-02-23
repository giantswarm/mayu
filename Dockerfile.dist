FROM ubuntu:wily

RUN apt-get update && \
    apt-get install -y dnsmasq-base git && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /etc/mayu /var/lib/mayu /usr/lib/mayu
COPY mayu mayuctl /usr/bin/
COPY tftproot /usr/lib/mayu/tftproot
COPY static_html /usr/lib/mayu/static_html
COPY template_snippets /usr/lib/mayu/template_snippets
COPY templates /usr/lib/mayu/templates
COPY config.yaml* /etc/mayu/

# enable if you want to add a post hook to github to store your cluster config
#RUN ssh-keyscan -H github.com > /etc/ssh/ssh_known_hosts

WORKDIR /usr/lib/mayu

RUN if [ ! -f /etc/mayu/config.yaml ]; then cp /etc/mayu/config.yaml.dist /etc/mayu/config.yaml; fi

ENTRYPOINT ["mayu", "--cluster-directory=/var/lib/mayu"]
CMD ["-v=12"]
