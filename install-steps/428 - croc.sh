#!/bin/bash

source common.sh


# P2P file transfer program, it's really magical! Go-port of "Magic Wormhole"

tempInstallDir="/tmp/croc-install"

mkdir "$tempInstallDir" && cd "$tempInstallDir"
curl -fsSL -o croc.deb \
	https://function61.com/app-dl/api/github.com/schollz/croc/latest_releases_asset/croc_%2A_Linux-64bit.deb
dpkg -i croc.deb
rm -rf "$tempInstallDir"

#versioncommand: croc --version
