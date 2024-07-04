#!/bin/bash

source common.sh


# Read logs from stdin and prints them back to stdout, but prettier.

curl -fsSL 'https://function61.com/app-dl/api/github.com/humanlogio/humanlog/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64' \
	| tar -C /usr/bin/ --wildcards 'humanlog' -xzf -

#versioncommand: humanlog --version | cut -d ' ' -f3
