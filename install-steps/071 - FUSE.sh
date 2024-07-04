#!/bin/bash

source common.sh


# A filesystem in which data and metadata are provided by an ordinary userspace process. 

# force-confnew to tell it to keep our custom `/etc/fuse.conf`:
# https://stackoverflow.com/a/7157196
DEBIAN_FRONTEND=noninteractive apt -o DPkg::Options::="--force-confnew" install -y fuse

#versioncommand: fusermount --version | cut -d' ' -f3
