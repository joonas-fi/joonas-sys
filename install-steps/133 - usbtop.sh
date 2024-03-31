#!/bin/bash

source common.sh


# usbutils = lsusb
apt install -y usbtop

#versioncommand: apt show usbtop | grep Version: | cut -d' ' -f2
