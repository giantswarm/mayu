#!/bin/bash

# gather values for templates
monsho_network=$(curl -g --unix-socket /var/run/docker.sock 'http:/networks?filters={"type":{"custom":true}}' \
  | jq '.[] | select(.Name == "monsho_default")')

subnet=$(echo "$monsho_network" \
  | jq -r '.IPAM.Config[].Subnet')

ip_addr_gateway=$(echo "$monsho_network" \
  | jq -r '.IPAM.Config[].Gateway | match("(.*)/") | .captures[0].string')

bridge_ifs="br-"$(echo "$monsho_network" \
  | jq -r '.Id[0:12]')

ip_addr_mayu=$(ip -family inet -oneline addr show up primary scope global dev eth0 \
  | while read num dev fam addr rest; do echo ${addr%/*}; done)

ip_addr_mayu=${ip_addr_mayu[0]}

# allow qemu to access our bridge
if ! grep -q "allow $bridge_ifs" /etc/qemu/bridge.conf ; then
  echo "allow $bridge_ifs" >> /etc/qemu/bridge.conf
fi

# fill template with values
sigil -f /etc/mayu/config.yaml.tmpl \
  subnet="$subnet" \
  ip_addr_mayu="$ip_addr_mayu" \
  ip_addr_gateway="$ip_addr_gateway" \
  > /etc/mayu/config.yaml


# FIXME! provide `gateway` key in mayu/config.yaml instead of these hacks:

sed s/%%ip_addr_gateway%%/$ip_addr_gateway/g /usr/lib/mayu/templates/last_stage_cloudconfig.yaml.tmpl \
  > /usr/lib/mayu/templates/last_stage_cloudconfig.yaml

sed s/%%ip_addr_gateway%%/$ip_addr_gateway/g /usr/lib/mayu/template_snippets/temp/net_bridge.yaml.tmpl \
  > /usr/lib/mayu/template_snippets/cloudconfig/net_bridge.yaml

sed s/%%ip_addr_gateway%%/$ip_addr_gateway/g /usr/lib/mayu/template_snippets/temp/net_bridge_dhcp.yaml.tmpl \
  > /usr/lib/mayu/template_snippets/cloudconfig/net_bridge_dhcp.yaml


exec mayu --cluster-directory=/var/lib/mayu "$@"
