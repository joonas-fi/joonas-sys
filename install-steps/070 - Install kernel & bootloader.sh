#!/bin/bash

source common.sh


# generic = not lowlatency or some other specific
# noninteractive because GRUB complains about bootsector
# DEBIAN_FRONTEND=noninteractive apt install -y grub-pc linux-image-generic

# hmm maybe we don't need GRUB
DEBIAN_FRONTEND=noninteractive apt install -y linux-image-generic

# have mount point ready for ESP
mkdir -p /boot/efi
