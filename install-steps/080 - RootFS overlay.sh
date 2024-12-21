#!/bin/bash

source common.sh


# programmatic steps related to having / be overlayfs that redirects writes to special persistence place
#
# most of the important things are already done in scripts in our overrides/etc/initramfs-tools/


# instruct pre-boot environment to have overlay kernel module loaded
# we could have this as static file, but then upstream changes would get overwritten
echo "overlay" >> /etc/initramfs-tools/modules

# we should run "$ update-initramfs ..." here, but as an optimization we have disabled it
# (so multiple steps don't trigger it) and we'll only run it once at the end of the process.
