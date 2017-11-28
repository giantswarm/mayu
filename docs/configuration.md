# Mayu Configuration

Here we provide more detailed documentation about configuring Mayu. By
default TLS is enabled when communicating with `mayu` over network. If your
setup does not provide or rely on TLS for whatever reasons, you can set
`--no-tls`. The corresponding flag for `mayuctl` is `--no-tls`.
Check [mayuctl](mayuctl.md) for more information about the client.

## File Tree

```nohighlight
.
|-- mayu                              - the mayu executable
|-- config.yaml.dist                  - mayu configuration file template
|-- templates
|   |-- dnsmasq_template.conf         - template file used to generate the dnsmasq configuration
|   |-- ignition.yaml.yaml            - template used to generate the ignition
|   |-- snippets                 - directory containing some template snippets used in the  ignition template
|   |   |-- net_bond.yaml
|   |   |-- net_singlenic.yaml
|   |   |-- extra.yaml
`-- tftproot
    `-- undionly.kpxe                 - ipxe pxe image
    `-- ipxe.efi                      - ipxe pxe image for UEFI enabled hosts
```

For a new environment to be configured, there are three main files that might
have to be adapted: `config.yaml`, `ignition.yaml` and one of the
 snippets `extra.yaml`, `net_bond.yaml` or `net_singlenic.yaml`.


## `/etc/mayu/config.yaml`

The very first thing to do is to copy `config.yaml.dist` to
`/etc/mayu/config.yaml` and modify it regarding your needs. The initial
section configures the network, profiles for the machines and the versions
of the software that should be installed via Yochu.

### Default Container Linux Version

To successfully run Mayu you need to specify a default Container Linux version. This version is used to bootstrap
machine. So whenever a new machine starts this Container Linux version is used to install Container Linux on the disk of
the machine. You can also specify other Container Linux versions within profiles or single machines that overwrite
this default value.

Most importantly you also need to fetch the Container Linux image version. This is explained in the [Running Mayu](running.md) section.

```yaml
default_coreos_version: 1409.7.0
```

### Network

```yaml
network:
  pxe: true
  uefi: false
  pxe_interface: eth0
  machine_interface: eth0
  bind_addr: 10.0.3.251
  bootstrap_range:
    start: 10.0.3.10
    end: 10.0.3.30
  ip_range:
    start: 10.0.4.31
    end: 10.0.4.70
  dns: [8.8.8.8]
  ntp: [0.pool.ntp.org, 1.pool.ntp.org]
  router: 10.0.3.251
  subnet_size: 24
  subnet_gateway: 10.0.4.251
  network_model: singlenic
```

Here we have three less obvious settings: the `bootstrap_range` is used by the
DHCP server during the bootstrap procedure and the nodes only use it during the
installation. The `ip_range` is a range of addresses that will be statically
assigned to the cluster nodes. The `network_model` specifies which network
template snippet will be used.


`pxe_interface` is defining on which network interface mayu should listen for pxe and dhcp


`machine_interface` is defining the name for interface that will be used for configuring network if `network_model: singlenic` is used.

### Profiles

```yaml
profiles:
  - name: core
    quantity: 3
  - name: default
```

Each profile has a `name`, a `quantity`
(defines the number of cluster nodes that should have this profile). Name can be used for distinguishing machines in the ignition templates. Once all the profiles' quantities are matched 
(in this example that means we have 3 nodes with the profile core), mayu will assign
the profile "default" to the remaining nodes. Thus, profiles with a `quantity`
set are of higher priority than the default profile.

### Template Variables For Cloudconfig

```yaml
templates_env:
  users:
    - Key: ssh-rsa xxxxxxxxxxxxxxx
      Name: my_user
    - Key: ssh-rsa yyyyyyyyyyyyyyy
      Name: second_user
  mayu_https_endpoint: https://10.0.1.254:4080
  mayu_http_endpoint: http://10.0.1.254:4081
  mayu_api_ip: 10.0.1.254
```

These variables are used by the templates (most of them are directly injected
into the ignition file).

## Commandline flags

```
--v=12
--cluster-directory=/var/lib/mayu
--alsologtostderr
--etcd-quorum-size=3
--etcd-endpoint=https://127.0.0.1:2379
--images-cache-dir=/var/lib/mayu/images
--yochu-path=/var/lib/mayu/yochu
--log_dir=/tmp
```

### Certificates

Communication between `mayu` and `mayuctl` by default is TLS encrypted. For
that you need to provide certificates as follows. To disable tls
you can set `--no-tls` to `true`. Then no certificate needs to be
provided.

```
--no-tls=false
--tls-cert-file="./cert.pem"
--tls_key-file="./key.pem"
```

## `ignition.yaml`

This template is a vanilla
[ignition](https://coreos.com/ignition/docs/latest/) file with a
few additions to automatically deploy the few units, define the etcd3 with discovery url, confgure ssh keys for users` and configure the network.

## `templates/snippets/net_singlenic.yaml`

In the near future, the existence of multiple network template snippets will be
changed, so we'll focus on the singlenic template (used by the default
configuration) for now.

```yaml
{{define "net_singlenic"}}
networkd:
  units:
  - name: 10-nodhcp.network
    contents: |
      [Match]
      Name=*

      [Network]
      DHCP=no
  - name: 00-{{.ClusterNetwork.MachineInterface}}.network
    contents: |
      [Match]
      Name={{.ClusterNetwork.MachineInterface}}

      [Network]
      Address={{.Host.InternalAddr}}/{{.ClusterNetwork.SubnetSize}}
      Gateway={{.ClusterNetwork.SubnetGateway}}
      {{ range $server := .ClusterNetwork.DNS }}DNS={{ $server }}
      {{ end }}
      {{ range $server := .ClusterNetwork.NTP }}NTP={{ $server }}
      {{ end }}
{{end}}
```

This snippet will be merged into the ignition file, so the right
indentation must be taken into account. The Container Linux [network
configuration](https://coreos.com/os/docs/latest/network-config-with-networkd.html)
defines the
[systemd-networkd](http://www.freedesktop.org/software/systemd/man/systemd.network.html)
.network (and optionally .device) files used by each node.

In this example it just disables DHCP and configures the `machine_interface` with a
static IP address. The `machine_interface` is configured in mayu config.


