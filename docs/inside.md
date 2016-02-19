# Mayu Cluster Insides

Here we looking inside a deployed cluster. New nodes can be added to the
cluster at any time and the same bootstrap procedure will be applied. In this
example, we created a 4 node cluster:

```nohighlight
$ cd cluster
$ egrep -r 'Host|InternalAddr' */conf.json
004b27ed-692e-b32e-1f68-d89aff66c71b/conf.json:  "InternalAddr": "10.0.3.31",
004b27ed-692e-b32e-1f68-d89aff66c71b/conf.json:  "Hostname": "00006811af601fe8",
2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a/conf.json:  "InternalAddr": "10.0.3.33",
2843c49e-d1ba-6dd3-1320-d7cc82d8ea3a/conf.json:  "Hostname": "0000906eb12096e3",
7100c054-d2c9-e299-b669-e8bdb85f6904/conf.json:  "InternalAddr": "10.0.3.32",
7100c054-d2c9-e299-b669-e8bdb85f6904/conf.json:  "Hostname": "0000d71391dc5317",
aa1f18e1-f14f-2dd9-4fa0-dae7317c712c/conf.json:  "InternalAddr": "10.0.3.34",
aa1f18e1-f14f-2dd9-4fa0-dae7317c712c/conf.json:  "Hostname": "0000b1895b74c624",
```

```nohighlight
$ ssh core@10.0.3.31 fleetctl list-machines -l
Warning: Permanently added '10.0.3.31' (ED25519) to the list of known hosts.
MACHINE         IP    METADATA
00006811af601fe8e1d3f37902021ae0  10.0.3.31 rule-core=true
0000906eb12096e3d94b002c663943f9  10.0.3.33 rule-core=true
0000b1895b74c624a51bd3b94d3adf3c  10.0.3.34 rule-worker=true,stack-compute=true
0000d71391dc5317a0a1798d6bd5448f  10.0.3.32 rule-core=true
```

We can observe that the profile `core` was assigned to the first 3 nodes and
the 4th node got the `default` profile. We should also note that each node's
hostname is a substring of the node's `machine-id`.

How It Works? Let's start by analyzing the bootstrap process of a fresh node:

![mayu bootstrap sequence](image/bootstrap.png)

Adding a fresh node to the cluster consists of three parts:

* initial boot
* system installation
* final reboot to installed system

## Initial boot

The fresh node is by definition empty and boots over ethernet by default. It
sends a DHCP request for a `pxeclient`, which gets answered by the management
node (which acts a DHCP/PXE server) with PXE details to boot iPXE. The node
then pulls iPXE boot data from the PXE server via tftp. The node now send
another DHCP request, this time with the ipxe client and gets back an ipxe boot
path. The node then requests the kernel image and subsequently the initial root
directory over HTTP. With this the node can then boot into installation from
the initial root directory.

## System Installation

The initial root directory contains a version of CoreOS that is modified to run
one unit that in turn requests a first stage script and runs that. The script
is based on the `cloudconfig` that the management node holds for each server
respectively and assigns based on the serial number of the machine. It waits
for connectivity and then downloads a vanilla CoreOS stable image, which is
cached on the management node. This image gets installed by default to
`/dev/sda` configured through the above-mentioned `cloudconfig`. Finally the
node announces itself as installed to the management node and reboots.

## Final Reboot To Installed System

On the final reboot (as well as each next reboot, as long as nothing in the
mayu configuration gets changed) the node will again try to boot over
ethernet first. It sends three DHCP request, which get ignored by the PXE
server. The PXE server ingores these as long as the node is marked as
`installed`. The node then continues booting the previously installed CoreOS
from `/dev/sda`.
