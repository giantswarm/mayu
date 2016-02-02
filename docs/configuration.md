# Mayu Configuration

Here we provide more detailed documentation about configuring Mayu. By
default TLS is enabled when communicating with `mayu` over network. If your
setup does not provide or rely on TLS for whatever reasons, you can set
`no_secure` to `true` within your `config.yaml`. The corresponding flag for
`mayuctl` is `--no-secure`. Check [mayuctl](mayuctl.md) for more
information about the client.

### File Tree

```
.
|-- mayu                            - the mayu executable
|-- config.yaml.dist                  - mayu configuration file template
|-- images                            - directory containing the boot and installation images
|   |-- ...
|-- static_html
|   `-- infopusher                    - small node info pusher used during the installation process
|-- templates
|   |-- dnsmasq_template.conf         - template file used to generate the dnsmasq configuration
|   |-- first_stage_script.sh         - template used to generate the installation script
|   `-- last_stage_cloudconfig.yaml   - template used to generate the final cloud-config
|-- template_snippets                 - directory containing some template snippets used by the final cloud-config
|   |-- net_bond.yaml
|   `-- net_singlenic.yaml
`-- tftproot
    `-- undionly.kpxe                 - ipxe pxe image
```

For a new environment to be configured, there are three main files that might
have to be adapted: `config.yaml`, `last_stage_cloudconfig.yaml` and one of the
network snippets `net_bond.yaml` or `net_singlenic.yaml`.

---

### /etc/mayu/config.yaml

The very first thing to do is to copy `config.yaml.dist` to
`/etc/mayu/config.yaml` and modify it regarding your needs. The initial
section configures the paths to the templates, some binaries (ipxe and dnsmasq)
and some directories. The default settings match the distributed mayu file
tree.

```
tftproot: ./tftproot
static_html_path: ./static_html
ipxe: undionly.kpxe
first_stage_script: ./templates/first_stage_script.sh
last_stage_cloudconfig: ./templates/last_stage_cloudconfig.yaml
dnsmasq_template: ./templates/dnsmasq_template.conf
template_snippets: ./template_snippets
dnsmasq: ./dnsmasq
images_cache_dir: ./images
http_port: 4080
```

---

###### Certificates

Communication between `mayu` and `mayuctl` by default is TLS encrypted. For
that you need to provide certificates as follows. To disable this security
feature you can set `no_secure` to `true`. Then no certificate needs to be
provided.

```
no_secure: false
https_cert_file: "./cert.pem"
https_key_file: "./key.pem"
```

###### Network

```
network:
  pxe: true
  interface: bond0
  bind_addr: 10.0.3.251
  bootstrap_range:
    start: 10.0.3.10
    end: 10.0.3.30
  ip_range:
    start: 10.0.3.31
    end: 10.0.3.70
  dns: [8.8.8.8]
  router: 10.0.3.251
  network_model: singlenic
```

Here we have three less obvious settings: the `bootstrap_range` is used by the
DHCP server during the bootstrap procedure and the nodes only use it during the
installation. The `ip_range` is a range of addresses that will be statically
assigned to the cluster nodes. The `network_model` specifies which network
template snippet will be used.

---

###### Profiles

```
profiles:
  - name: core
    quantity: 3
    tags:
      - "rule-core=true"
  - name: default
    tags:
      - "rule-worker=true"
      - "stack-compute=true"
```

The final goal of a mayu-enabled deployment is a functional fleet cluster. To
be able to assign different roles to the different nodes, mayu employs a
mechanism of profile selection. Each profile has a `name`, a `quantity`
(defines the number of cluster nodes that should have this profile) and a list
of `tags` (the elements of this list will be directly mapped to fleet metadata
tags). Once all the profiles' quantities are matched (in this example that
means we have 3 nodes with the profile core), mayu will assign the profile
"default" to the remaining nodes. Thus, profiles with a `quantity` set are of
higher priority than the default profile.

---

###### Template Variables For Cloudconfig

```
templates_env:
  ssh_authorized_keys:
    - "ssh-rsa ..."
    - "ssh-rsa ..."
  yochu_localsubnet: "10.0.0.0/22"
  yochu_gateway: "10.0.3.251/32"
  yochu_private_registry: "registry.<cluster-name>.private.<domain>"
```

These variables are used by the templates (most of them are directly injected
into the final cloudconfig file).

### last_stage_cloudconfig.yaml

This template is a vanilla
[cloud-config](https://coreos.com/os/docs/latest/cloud-config.html) file with a
few additions to automatically deploy the Giant Swarm yochu, setup the
fleet metadata, define the etcd discovery url, update the `ssh_authorized_keys`
for the user `core` and configure the network.

### template_snippets/net_singlenic.yaml

In the near future, the existence of multiple network template snippets will be
changed, so we'll focus on the singlenic template (used by the default
configuration) for now.

```
{{define "net_singlenic"}}
  - name: systemd-networkd.service
    command: stop
  - name: 10-nodhcp.network
    runtime: false
    content: |
      [Match]
      Name=*

      [Network]
      DHCP=no
  - name: 00-{{.Host.ConnectedNIC}}.network
    runtime: false
    content: |
      [Match]
      Name={{.Host.ConnectedNIC}}

      [Network]
      Address={{.Host.InternalAddr}}/22
      Gateway={{.ClusterNetwork.Router}}
      DNS={{index .ClusterNetwork.DNS 0}}
  - name: down-interfaces.service
    command: start
    content: |
      [Service]
      Type=oneshot
      ExecStart=/usr/bin/ip link set {{.Host.ConnectedNIC}} down
      ExecStart=/usr/bin/ip addr flush dev {{.Host.ConnectedNIC}}
  - name: systemd-networkd.service
    command: restart
{{end}}
```

This snippet will be merged into the `cloud-config` file, so the right
indentation must be taken into account. The CoreOS [network
configuration](https://coreos.com/os/docs/latest/network-config-with-networkd.html)
defines the
[systemd-networkd](http://www.freedesktop.org/software/systemd/man/systemd.network.html)
.network (and optionally .device) files used by each node.

In this example it just disables DHCP and configures the `ConnectedNIC` with a
static IP address. The `ConnectedNIC` is aquired during installation time by
analyzing which interface is used to reach the default gateway.

