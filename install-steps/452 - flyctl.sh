#!/bin/bash

source common.sh


# IaaS platform

curl -fsSL \
	https://function61.com/app-dl/api/github.com/superfly/flyctl/latest_releases_asset/flyctl_%2A_Linux_x86_64.tar.gz \
	| tar -C /usr/bin/ -xzf -

#versioncommand: flyctl version
