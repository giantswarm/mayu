# Mayu Network Requirements

Machines need to be able to connect to Mayu on a few ports. Here is a short list on what is used.

```
PORT    PROTOCOL  DESCRIPTION
4080    TCP       default TLS/HTTP port to communicate the state of the machine and fetch binaries and scripts to provision machines
67      UDP       DHCP/BOOTP to let machines boot via PXE/iPXE
69      UDP/TCP   TFTP to ship images via PXE
```  
