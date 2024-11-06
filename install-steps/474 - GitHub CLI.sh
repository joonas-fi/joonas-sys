#!/bin/bash

source common.sh


# GitHub’s official command line tool

curl -fsSL 'https://function61.com/app-dl/api/github.com/cli/cli/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin/ --wildcards '*/bin/gh' --strip-components=2 -xzf -

#versioncommand: gh --version | grep -oP '(\d+\.\d+\.\d+)'
