#!/bin/bash

source common.sh


# generic = not lowlatency or some other specific
# noninteractive because GRUB complains about bootsector
# DEBIAN_FRONTEND=noninteractive apt install -y grub-pc linux-image-generic

# kexec-tools for fast reboots

DEBIAN_FRONTEND=noninteractive apt install -y linux-image-generic kexec-tools
