#!/bin/bash

source common.sh


# generic = not lowlatency or some other specific
# noninteractive because GRUB complains about bootsector
# DEBIAN_FRONTEND=noninteractive apt install -y grub-pc linux-image-generic

# kexec-tools for fast reboots

# weird options because we have overridden /etc/initramfs-tools/update-initramfs.conf
#   https://serverfault.com/a/593640
DEBIAN_FRONTEND=noninteractive apt install -y \
	-o Dpkg::Options::="--force-confdef" -o Dpkg::Options::="--force-confold" \
	linux-image-generic kexec-tools

# can't use uname --kernel-release because it operates on the running kernel, while we're most
# likely inside a build container
#versioncommand: readlink /boot/vmlinuz
