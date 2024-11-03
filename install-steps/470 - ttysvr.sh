#!/bin/bash

source common.sh


# Screensavers for your terminal

curl -fsSL 'https://function61.com/app-dl/api/github.com/cxreiff/ttysvr/latest_releases_asset/__autodetect__.tar.xz?os=linux&arch=amd64' \
	| tar -C /usr/bin --wildcards '*/ttysvr' --strip-components=1 -xJf -

#versioncommand: ttysvr --version | cut -d ' ' -f2
