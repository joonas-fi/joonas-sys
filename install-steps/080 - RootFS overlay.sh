#!/bin/bash

source common.sh


# programmatic steps related to having / be overlayfs that redirects writes to /persist/sys_N_diffs
#
# most of the important things are already done in scripts in our overrides/etc/initramfs-tools/


# instruct pre-boot environment to have overlay kernel module loaded
# we could have this as static file, but then upstream changes would get overwritten
echo "overlay" >> /etc/initramfs-tools/modules

# update needed after we modified the contents (later steps probably do this, but it'd be dirty to rely on it)
update-initramfs -u -k all
