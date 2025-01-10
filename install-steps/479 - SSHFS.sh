#!/bin/bash

source common.sh


# SSHFS is a FUSE-based file system client for mounting remote directories over a Secure Shell connection.
# https://wiki.archlinux.org/title/SSHFS

apt install -y sshfs

#versioncommand: apt show sshfs | grep Version: | cut -d' ' -f2
