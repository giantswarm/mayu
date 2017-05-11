#!/bin/bash
# -*- mode:shell-script; indent-tabs-mode:nil; sh-basic-offset:2 -*-

which wget >/dev/null || exit 1
which curl >/dev/null || exit 1
which gpg >/dev/null || exit 1

#make sure we have connectivity

while [[ -z $(curl -k -s {{.MayuURL}}) ]] ; do sleep 1;  done

wget -O /tmp/helper {{.HostInfoHelperURL}}
chmod +x /tmp/helper

if [[ -n "{{.CloudConfigURL}}" ]]; then
  /tmp/helper -post-url={{.CloudConfigURL}} > /tmp/future-cloud-config.yaml
fi
if [[ -n "{{.IgnitionConfigURL}}" ]]; then
  /tmp/helper -post-url={{.IgnitionConfigURL}} > /tmp/future-ignition-config.json
fi
C=$(cat /proc/cmdline | tr ' ' '\n' | grep maybe-install-coreos= | awk -F=  '{print $2}')

serial=$(cat /sys/devices/virtual/dmi/id/{product,chassis,board}_serial | tr  ' ' '-' | sed '/^\s*$/d' | head -1)

DEVICE=""
CLOUDINIT=""
CHANNEL_ID=""
IGNITION=""


coreos_install() {
  set -e -o pipefail

  umask 077

  # Device is required, must not be a partition, must be writable
  if [[ -z "${DEVICE}" ]]; then
    echo "$0: No target block device provided, -d is required." >&2
    exit 1
  fi

  if ! [[ $(lsblk -n -d -o TYPE "${DEVICE}") =~ ^(disk|loop|lvm)$ ]]; then
    echo "$0: Target block device (${DEVICE}) is not a full disk." >&2
    exit 1
  fi

  if [[ ! -w "${DEVICE}" ]]; then
    echo "$0: Target block device (${DEVICE}) is not writable (are you root?)" >&2
    exit 1
  fi

  if [[ -n "${CLOUDINIT}" ]]; then
    if [[ ! -f "${CLOUDINIT}" ]]; then
      echo "$0: Cloud config file (${CLOUDINIT}) does not exist." >&2
      exit 1
    fi

    if type -P coreos-cloudinit >/dev/null; then
      if ! coreos-cloudinit -from-file="${CLOUDINIT}" -validate; then
        echo "$0: Cloud config file (${CLOUDINIT}) is not valid." >&2
        exit 1
      fi
    else
      echo "$0: coreos-cloudinit not found. Could not validate config. Continuing..." >&2
    fi
  fi

  if [[ -n "${IGNITION}" ]]; then
      if [[ ! -f "${IGNITION}" ]]; then
          echo "$0: Ignition config file (${IGNITION}) does not exist." >&2
          exit 1
      fi
  fi

  IMAGE_URL="{{.InstallImageURL}}"

  # Pre-flight checks pass, lets get this party started!
  WORKDIR=$(mktemp --tmpdir -d coreos-install.XXXXXXXXXX)
  trap "rm -rf '${WORKDIR}'" EXIT

  echo "Downloading, writing and verifying ${IMAGE_NAME}..."
  declare -a EEND
  if ! wget --inet4-only --no-verbose -O - "${IMAGE_URL}" \
    | bunzip2 --stdout >"${DEVICE}"
then
  EEND=(${PIPESTATUS[@]})
  [ ${EEND[0]} -ne 0 ] && echo "${EEND[0]}: Download of ${IMAGE_NAME} did not complete" >&2
  [ ${EEND[1]} -ne 0 ] && echo "${EEND[1]}: Cannot expand ${IMAGE_NAME} to ${DEVICE}" >&2
  wipefs --all --backup "${DEVICE}"
  exit 1
fi

sync
sleep 1
udevadm settle
blockdev --rereadpt "${DEVICE}" || partprobe ${DEVICE}
sleep 1

if [[ -n "${CLOUDINIT}" ]] || [[ -n "${COPY_NET}" ]]; then
  # The ROOT partition should be #9 but make no assumptions here!
  # Also don't mount by label directly in case other devices conflict.
  ROOT_DEV=$(blkid -t "LABEL=ROOT" -o device "${DEVICE}"*)

  if [[ -z "${ROOT_DEV}" ]]; then
    echo "Unable to find new ROOT partition on ${DEVICE}" >&2
    exit 1
  fi

  mkdir -p "${WORKDIR}/rootfs"
  case $(blkid -t "LABEL=ROOT" -o value -s TYPE "${ROOT_DEV}") in
    "btrfs") mount -t btrfs -o subvol=root "${ROOT_DEV}" "${WORKDIR}/rootfs" ;;
    *)       mount "${ROOT_DEV}" "${WORKDIR}/rootfs" ;;
  esac
  trap "umount '${WORKDIR}/rootfs' && rm -rf '${WORKDIR}'" EXIT

  if [[ -n "${CLOUDINIT}" ]]; then
    echo "Installing cloud-config..."
    mkdir -p "${WORKDIR}/rootfs/var/lib/coreos-install"
    cp "${CLOUDINIT}" "${WORKDIR}/rootfs/var/lib/coreos-install/user_data"
  fi

  if [[ -n "${COPY_NET}" ]]; then
    echo "Copying network units to root partition."
    # Copy the entire directory, do not overwrite anything that might exist there, keep permissions, and copy the resolve.conf link as a file.
    cp --recursive --no-clobber --preserve --dereference /run/systemd/network/* "${WORKDIR}/rootfs/etc/systemd/network"
  fi
  echo "{{.MachineID}}" > "${WORKDIR}/rootfs/etc/machine-id"

  echo "SERIAL=$serial" > "${WORKDIR}/rootfs/etc/mayu-env"
  echo "MAYU_VERSION={{.MayuVersion}}" >> "${WORKDIR}/rootfs/etc/mayu-env"

  umount "${WORKDIR}/rootfs"
fi

if [[ -n "${IGNITION}" ]]; then
    # The OEM partition should be #3 but make no assumptions here!
    # Also don't mount by label directly in case other devices conflict.
    OEM_DEV=$(blkid -t "LABEL=OEM" -o device "${DEVICE}"*)

    if [[ -z "${OEM_DEV}" ]]; then
      echo "Unable to find new OEM partition on ${DEVICE}" >&2
      exit 1
    fi

    mkdir -p "${WORKDIR}/oemfs"
    mount "${OEM_DEV}" "${WORKDIR}/oemfs"
    trap "umount '${WORKDIR}/oemfs' && rm -rf '${WORKDIR}'" EXIT

    echo "Installing Ignition config ${IGNITION}..."
    cp "${IGNITION}" "${WORKDIR}/oemfs/coreos-install.json"
    echo  "set linux_append=\"coreos.config.url=oem:///coreos-install.json"\" > "${WORKDIR}/oemfs/grub.cfg"

    umount "${WORKDIR}/oemfs"
    trap "rm -rf '${WORKDIR}'" EXIT
fi

rm -rf "${WORKDIR}"
trap - EXIT

echo "Success! CoreOS ${CHANNEL_ID} ${VERSION_ID}${OEM_ID:+ (${OEM_ID})} is installed on ${DEVICE}"


}



case "$C" in
  "alpha"|"beta"|"stable" )

    echo "installing $C" to disk;

    CHANNEL_ID=$C
    DEVICE=/dev/sda
    if [[ -n "{{.CloudConfigURL}}" ]]; then
      CLOUDINIT=/tmp/future-cloud-config.yaml
    fi
    if [[ -n "{{.IgnitionConfigURL}}" ]]; then
      IGNITION=/tmp/future-ignition-config.json
    fi
    set -e
    coreos_install

    curl -X PUT --header "Content-Length: 0" "$(echo {{.SetInstalledURL}} | sed "s/__SERIAL__/$serial/")" && reboot

    ;;
  *)
    echo "not installing coreos to disk";
    ;;
esac
