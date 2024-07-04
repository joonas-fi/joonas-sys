#!/bin/bash

source common.sh


# IaaS platform

curl -fsSL \
	"https://function61.com/app-dl/api/github.com/superfly/flyctl/latest_releases_asset/__autodetect__.tar.gz?os=linux&arch=amd64" \
	| tar -C /usr/bin/ -xzf -

#versioncommand: flyctl version | cut -d' ' -f 2
