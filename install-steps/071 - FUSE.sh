#!/bin/bash

source common.sh

# the installer will complain if it would try to overwrite our config file.
# we'll restore this after installation is complete
rm /etc/fuse.conf

DEBIAN_FRONTEND=noninteractive apt install -y fuse

cp "${repodir}/overrides/etc/fuse.conf" /etc/fuse.conf

#versioncommand: fusermount --version | cut -d' ' -f3
