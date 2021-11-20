#!/bin/bash

source common.sh


# snap/snapcraft is Docker-like but mainly focused for GUI apps
apt install -qy snapd

#versioncommand: snap --version | head -1
