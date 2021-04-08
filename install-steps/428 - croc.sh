#!/bin/bash

source common.sh


# P2P file transfer program, it's really magical! Go-port of "Magic Wormhole"

tempInstallDir="/tmp/croc-install"

mkdir "$tempInstallDir" && cd "$tempInstallDir"
curl -fsSL -o croc.deb https://github.com/schollz/croc/releases/download/v8.6.7/croc_8.6.7_Linux-64bit.deb
dpkg -i croc.deb
rm -rf "$tempInstallDir"

#versioncommand: croc --version
