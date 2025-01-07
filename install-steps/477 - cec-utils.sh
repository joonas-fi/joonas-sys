#!/bin/bash

source common.sh


# cec-client for HDMI-CEC: https://en.wikipedia.org/wiki/Consumer_Electronics_Control
apt install -y cec-utils

#versioncommand: apt show cec-utils | grep Version: | cut -d' ' -f2
