#!/bin/bash

source common.sh

# somewhere along the process AMD GPU drivers seem to get automatically pulled in..
# so do nouveau drivers (Nvidia open source drivers).

# we've to add Nvidia proprietary driver, because nouveau doesn't support video decoding acceleration for GeForce GT 1030
apt install -y nvidia-driver-460

# info utility for video acceleration debug
apt install -y vainfo
