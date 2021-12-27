#!/bin/bash

source common.sh


# numlock state on at boot
apt install -y numlockx

#versioncommand: apt show numlockx | grep Version: | cut -d' ' -f2
