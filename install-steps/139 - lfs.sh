#!/bin/bash

source common.sh


# A linux utility listing your filesystems.

mkdir /tmp/lfs
cd /tmp/lfs
wget https://function61.com/app-dl/api/github.com/Canop/lfs/latest_releases_asset/lfs_%2A.zip
unzip lfs_*.zip
mv build/x86_64-linux/lfs /usr/bin/
cd /
rm -rf /tmp/lfs

#versioncommand: lfs --version | cut -d' ' -f2
