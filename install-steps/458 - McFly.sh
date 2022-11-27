#!/bin/bash

source common.sh


# fly through your shell history

curl -fsSL https://function61.com/app-dl/api/github.com/cantino/mcfly/latest_releases_asset/mcfly-%2A-x86_64-unknown-linux-musl.tar.gz \
	| tar -C /usr/bin -xzf -

#versioncommand: mcfly --version | cut -d ' ' -f2

