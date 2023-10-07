#!/bin/bash

source common.sh


# A modern watch command. Time machine and pager etc.

curl -fsSL "https://function61.com/app-dl/api/github.com/sachaos/viddy/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64" \
	| tar -C /usr/bin -xzf - viddy

#versioncommand: viddy --version | cut -d ' ' -f3

