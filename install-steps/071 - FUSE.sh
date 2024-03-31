#!/bin/bash

source common.sh


# A filesystem in which data and metadata are provided by an ordinary userspace process. 

DEBIAN_FRONTEND=noninteractive apt install -y fuse

#versioncommand: fusermount --version | cut -d' ' -f3
