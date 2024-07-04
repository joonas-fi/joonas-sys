#!/bin/bash

source common.sh


# Linux kernel version specific tools (contains among others USB/IP software: https://wiki.archlinux.org/title/USB/IP )

DEBIAN_FRONTEND=noninteractive apt install -y linux-tools-generic

#versioncommand: apt show linux-tools-generic | grep Version: | cut -d' ' -f2
