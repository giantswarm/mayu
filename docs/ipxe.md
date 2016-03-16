# iPXE Setup

In order to install machines `mayu` uses iPXE to ship operating system
binaries. Hosts that are going to be managed with `mayu` will retrieve
iPXE scripts as response to their DHCP request. For more information about this
process take a look at [Mayu Cluster Insides](inside.md).

__Note: If you don't know how to use iPXE you can follow the [official
instructions](http://ipxe.org/start#quick_start).__

## TLS Support

When using `mayu` in TLS mode make sure that you use an iPXE version which
was compiled with [`DOWNLOAD_PROTO_HTTPS`](http://ipxe.org/buildcfg/download_proto_https)
support.

Also make sure when providing a custom SSL certificate that you need to follow
the [`cryptography`](http://ipxe.org/crypto) instuctions of iPXE.
