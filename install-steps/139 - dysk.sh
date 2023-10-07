#!/bin/bash

source common.sh


# A linux utility listing your filesystems.

mkdir -p /tmp/dysk
cd /tmp/dysk
wget https://function61.com/app-dl/api/github.com/Canop/dysk/latest_releases_asset/dysk_%2A.zip
unzip dysk_*.zip
mv build/x86_64-linux/dysk /usr/bin/
cd /
rm -rf /tmp/dysk

#versioncommand: dysk --version | cut -d' ' -f2
