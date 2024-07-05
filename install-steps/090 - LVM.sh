#!/bin/bash

source common.sh


# Logical Volume Management. snapshots, full-disk encryption etc.
apt install -y lvm2 cryptsetup

#versioncommand: apt show lvm2 | grep Version: | cut -d' ' -f2
