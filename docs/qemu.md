# Start KVMs inside CoreOS

If you would like to start CoreOS via KVM inside your physical machines then Mayu can also serve the necessary assets for you.

Btw if you are looking for a way to test Mayu with Qemu. We have created [Onsho](https://github.com/giantswarm/onsho) to reproduce our datacenter setup on a laptop.

## Prepare Mayu

Within a release (or after running `make bin-dist`) you will find a script called `./fetch-coreos-qemu-image.sh`. This image will download the PXE image and kernel but extract the `/usr` filesystem and put it in a folder that can be served by Mayu.

Note: You need to install the CoreOS image signing key to be able to verify the downloads. See https://coreos.com/os/docs/latest/verify-images.html

To fetch CoreOS `1122.2.0` you can run:
```
./fetch-coreos-qemu-image 1122.2.0
```

Or if you prefer the alpha channel use:
```
./fetch-coreos-qemu-image 1068.0.0 alpha
```

This will download the image into a folder called `./images/qemu/<coreos-version>`

## Create a container

This is an example how to create a VM inside of CoreOS. To have some tooling available it is easier to start KVM from within a container. So you need a Dockerfile.

```
FROM fedora:latest

RUN dnf -y update && \
    dnf install -y net-tools libattr libattr-devel xfsprogs bridge-utils qemu-kvm  qemu-system-x86 qemu-img && \
    dnf clean all

ADD run.sh /run.sh
ADD cloudconfig.yaml /usr/code/cloudconfig/openstack/latest/user_data

RUN mkdir -p /usr/code/{rootfs,images}

ENTRYPOINT ["/run.sh"]
```

The entrypoint creates a rootfs and starts the actual qemu process to start the virtual machine. This assumes that you have a bridge called `br0` on the host.

```
#!/bin/bash

set -eu

echo "allow br0" > /etc/qemu/bridge.conf

ROOTFS=/usr/code/rootfs/rootfs.img
KERNEL=/usr/code/images/coreos_production_qemu.vmlinuz
USRFS=/usr/code/images/coreos_production_qemu_usr_image.squashfs
MAC_ADDRESS=$(printf '%02X:%02X:%02X:%02X:%02X:%02X\n' $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)))

if [ ! -f $ROOTFS ]; then
    truncate -s 4G $ROOTFS
    mkfs.xfs $ROOTFS
fi

exec /usr/bin/qemu-system-x86_64 \
    -nographic \
    -machine accel=kvm -cpu host -smp 4 \
    -m 1024 \
    -enable-kvm \
    \
    -net bridge,br=$BRIDGE_NETWORK,vlan=0,helper=/usr/libexec/qemu-bridge-helper \
    -net nic,vlan=0,model=virtio,macaddr=$MAC_ADDRESS \
    \
    -fsdev local,id=conf,security_model=none,readonly,path=/usr/code/cloudconfig \
    -device virtio-9p-pci,fsdev=conf,mount_tag=config-2 \
    \
    -drive if=virtio,file=$USRFS,format=raw,serial=usr.readonly \
    -drive if=virtio,file=$ROOTFS,format=raw,discard=on,serial=rootfs \
    \
    -device sga \
    -serial mon:stdio \
    \
    -kernel $KERNEL \
    -append "console=ttyS0 root=/dev/disk/by-id/virtio-rootfs rootflags=rw mount.usr=/dev/disk/by-id/virtio-usr.readonly mount.usrflags=ro"
```

Don't forget to add your own cloudconfig.yaml for your VM. Then build the container image: `docker build -t giantswarm/coreos-qemu .`.

## Fetch the image

Now you just have to fetch the assets from Mayu and start the VM on the host itself. Fetching can be done via a cloudconfig unit on the host. So you need to include this snippet in your `./templates/last_stage_cloudconfig.yaml`.

```
coreos:
  units:
  - name: fetch-qemu-images.service
    command: start
    enable: true
    content: |
      [Unit]
      Description=Fetch qemu images from Mayu
      Wants=network-online.target
      After=network-online.target

      [Service]
      Type=oneshot
      Environment="IMAGE_DIR=/home/core/images"
      Environment="KERNEL=coreos_production_qemu.vmlinuz"
      Environment="USRFS=coreos_production_qemu_usr_image.squashfs"
      ExecStartPre=/bin/mkdir -p ${IMAGE_DIR}
      ExecStartPre=/usr/bin/wget {{index .TemplatesEnv "mayu_http_endpoint"}}/images/{{.Host.Serial}}/qemu/${KERNEL} -O ${IMAGE_DIR}/${KERNEL}
      ExecStartPre=/usr/bin/wget {{index .TemplatesEnv "mayu_http_endpoint"}}/images/{{.Host.Serial}}/qemu/${KERNEL}.sha256 -O ${IMAGE_DIR}/${KERNEL}.sha256
      ExecStartPre=/usr/bin/wget {{index .TemplatesEnv "mayu_http_endpoint"}}/images/{{.Host.Serial}}/qemu/${USRFS} -O ${IMAGE_DIR}/${USRFS}
      ExecStartPre=/usr/bin/wget {{index .TemplatesEnv "mayu_http_endpoint"}}/images/{{.Host.Serial}}/qemu/${USRFS}.sha256 -O ${IMAGE_DIR}/${USRFS}.sha256
      ExecStart=/bin/bash -c "cd ${IMAGE_DIR} && sha256sum -c ${USRFS}.sha256 && sha256sum -c ${KERNEL}.sha256"

      [Install]
      WantedBy=multi-user.target
```

## Start the VM

Finally you can start the virtual machine by running the container:

```
mkdir -p /home/core/vms/foo
docker run -ti
    --privileged \
    --net=host \
    -v $(pwd)/images:/usr/code/images \
    -v /home/core/vms/foo/:/usr/code/rootfs/ \
    giantswarm/coreos-qemu
```
