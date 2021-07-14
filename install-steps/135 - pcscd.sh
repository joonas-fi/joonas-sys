#!/bin/bash

source common.sh


# Smart Card daemon (so we can access YubiKeys etc.)
apt install -y pcscd

#versioncommand: apt show pcscd | grep Version: | cut -d' ' -f2
