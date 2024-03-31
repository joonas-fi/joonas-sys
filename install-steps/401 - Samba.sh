#!/bin/bash

source common.sh

apt install --no-install-recommends -y samba

# we only install Samba for QEMU Samba-based network file share support
# (it runs its own instance: doesn't need the background service)
systemctl disable smbd

#versioncommand: samba --version
