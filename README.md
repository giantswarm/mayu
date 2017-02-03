# Mayu

[![CircleCI](https://circleci.com/gh/giantswarm/mayu.svg?&style=shield)](https://circleci.com/gh/giantswarm/mayu)
[![](https://godoc.org/github.com/giantswarm/mayu?status.svg)](http://godoc.org/github.com/giantswarm/mayu) [![](https://img.shields.io/docker/pulls/giantswarm/mayu.svg)](http://hub.docker.com/giantswarm/mayu)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/mayu)](https://goreportcard.com/report/github.com/giantswarm/mayu)
[![IRC Channel](https://img.shields.io/badge/irc-%23giantswarm-blue.svg)](https://kiwiirc.com/client/irc.freenode.net/#giantswarm)

Mayu provides a set of mechanisms to bootstrap PXE-enabled bare metal nodes
that must follow a specific configuration with CoreOS. It sets up fleet
meta-data, and patched versions of fleet, etcd, and docker when using
[Yochu](https://github.com/giantswarm/yochu).

## Prerequisites

Mayu requires some basic configuration and layer 2 connectivity to the rest
of the nodes. Usually the cluster’s management node is used for this. The
management node acts as a PXE server and should support three kinds of requests
from the rest of the nodes: PXE, DHCP, and bootp. The rest of the nodes should
be configured to boot via ethernet by default and share a network segment with
the management node, so they get the PXE boot data from the management node on
DHCP request.

You need `dnsmasq` (>= 2.75) installed. Make sure dnsmasq is not running via an
init script or systemd after you installed it, as mayu starts its own
dnsmasq. In case you don't want to care about this, you maybe want to use the
docker image, where the dependency is built in. This eases the setup.

Further, mayu will keep track of all changes to the cluster by making git
commits. This is a feature for production systems and requires `git` (> 1.7.4)
installed. Use the `-no-git` flag when starting mayu to turn this feature off.

Developing Mayu requires the following tools to be installed.

 * `wget`
 * `go-bindata`
 * `cpio`

## Getting Mayu

Download the latest release from here: https://github.com/giantswarm/mayu/releases/latest

Clone the latest git repository version from here: https://github.com/giantswarm/mayu.git

Get the latest docker image from here: https://hub.docker.com/r/giantswarm/mayu/

## Running Mayu

Configuring Mayu is explained in [docs/configuration.md](docs/configuration.md). After configuration have
a look at [docs/running.md](docs/running.md) on how to start Mayu.

## Further Steps

Check more detailed documentation: [docs](docs)

Check code documentation: [godoc](https://godoc.org/github.com/giantswarm/mayu)

## Future Development

- Future directions/vision

## Contact

- Mailing list: [giantswarm](https://groups.google.com/forum/#!forum/giantswarm)
- IRC: #[giantswarm](irc://irc.freenode.org:6667/#giantswarm) on freenode.org
- Bugs: [issues](https://github.com/giantswarm/mayu/issues)

## Contributing & Reporting Bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## License

Mayu is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

## Origin of the Name

`mayu` (まゆ[繭] pronounced "mah-yoo") is Japanese for cocoon.
