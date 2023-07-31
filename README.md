[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/mayu/tree/master.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/mayu/tree/master)
[![](https://pkg.go.dev/badge/github.com/giantswarm/mayu)](https://pkg.go.dev/github.com/giantswarm/mayu)
[![](https://img.shields.io/docker/pulls/giantswarm/mayu.svg)](http://hub.docker.com/giantswarm/mayu)
[![Go Report Card](https://goreportcard.com/badge/github.com/giantswarm/mayu)](https://goreportcard.com/report/github.com/giantswarm/mayu)

# Mayu
Mayu provides a set of mechanisms to bootstrap PXE-enabled bare metal nodes
that must follow a specific configuration with Container Linux. 

## Prerequisites

Mayu requires some basic configuration and layer 2 connectivity to the rest
of the nodes. Usually the cluster’s management node is used for this. The
management node acts as a PXE server and should support three kinds of requests
from the rest of the nodes: PXE, DHCP, and bootp. The rest of the nodes should
be configured to boot via ethernet by default and share a network segment with
the management node, so they get the PXE boot data from the management node on
DHCP request.

Developing Mayu requires the following tools to be installed.

* `wget`
* `go-bindata`
* `cpio`

## Getting Mayu

Get the latest Docker image here: https://quay.io/repository/giantswarm/mayu

Clone the latest git repository version from here: https://github.com/giantswarm/mayu.git

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
