#!/bin/bash

source common.sh

# the below command does nothing if "update_initramfs=no"
echo "update_initramfs=yes" > /etc/initramfs-tools/update-initramfs.conf

update-initramfs -c -k all
