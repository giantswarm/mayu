# Security Overview

Mayu uses `Transport Layer Security <https://en.wikipedia.org/wiki/Transport_Layer_Security>` (TLS)
to encrypt connections between `mayu` and `mayuctl` and other clients like curl.

When using `mayu` in TLS mode make sure that you use an iPXE image, which
was compiled with [`DOWNLOAD_PROTO_HTTPS`](http://ipxe.org/buildcfg/download_proto_https)
support.

Further, when providing a custom SSL certificate, you should follow
the [`cryptography`](http://ipxe.org/crypto) instuctions of iPXE.

## Risks

Note that the above-mentioned TLS only provides encryption, not authentication.

Mayu communicates over following protocols: DHCP, TFTP, iPXE, and HTTP/HTTPS.
Currently, there is two general security notices around these protocols,
which you can find at the end of this section.

We recommend to run `mayu` within a separate network
with limited access by non-authorized users. This way the lack of
authentification as well as the general protocol issues are less critical.

### iPXE

See  http://security.stackexchange.com/questions/64915/what-are-the-biggest-security-concerns-on-pxe

> The basic PXE process:
>
> - Computer makes a DHCP request
> - DHCP server responds with address and PXE parameters
> - Computer downloads boot image using TFTP over UDP
>
> The obvious attacks are a rogue DHCP server responding with bad data (and thus
> hijacking the boot process) and a rogue TFTP server blindly injecting forged
> packets (hijacking or corrupting the boot image).
>
> UEFI secure boot can be used to prevent hijacking, but a rogue DHCP or TFTP
> server can still prevent booting by ensuring the computer receives a corrupted
> boot image.

### TFTP

See  https://technet.microsoft.com/en-us/library/cc754605.aspx

> The TFTP protocol does not support any authentication or encryption mechanism,
> and as such can introduce a security risk when present. Installing the TFTP
> client is not recommended for systems that access the Internet.
