#!/bin/bash

source common.sh


echo "work" > /etc/hostname

# the .network files were specified in overrides/
# we could do network config with /etc/network but I guess systemd-networkd has advantages?

# https://wiki.archlinux.org/index.php/systemd-networkd
# for some reason the network configuration daemon is not up by default
systemctl enable systemd-networkd
