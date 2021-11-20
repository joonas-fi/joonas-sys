#!/bin/bash

source common.sh


# apparently we need to download 42 MB of packages (including GPG, pinentry and gstreamer!!!)
# to be able to get stuff from PPA repositories
apt install -qy software-properties-common
