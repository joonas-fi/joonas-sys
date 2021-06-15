#!/bin/bash

source common.sh


# CLI filesystem navigator

curl -fsSL \
	https://function61.com/app-dl/api/github.com/gokcehan/lf/latest_releases_asset/lf-linux-amd64.tar.gz \
	| tar -C /usr/bin/ -xzf -

#versioncommand: lf --version
