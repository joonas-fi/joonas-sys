#!/bin/bash

source common.sh


# binutils = strings
apt install -y binutils

#versioncommand: apt show binutils | grep Version: | cut -d' ' -f2
