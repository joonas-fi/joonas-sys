#!/bin/bash

source common.sh


# usbutils = lsusb
apt install -y usbutils

#versioncommand: apt show usbutils | grep Version: | cut -d' ' -f2
