#!/bin/bash

source common.sh


# Android Debug Bridge + fastboot
apt install -y adb fastboot

#versioncommand: apt show adb | grep Version: | cut -d' ' -f2
