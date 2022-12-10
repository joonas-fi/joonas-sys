#!/bin/bash

source common.sh


# For managing firmware updates

apt install -y fwupd

#versioncommand: apt show fwupd | grep Version: | cut -d' ' -f2
