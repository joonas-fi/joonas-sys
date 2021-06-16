#!/bin/bash

source common.sh

# --wildcards because the binary's inside a dynamically named directory
curl -fsSL \
	https://function61.com/app-dl/api/github.com/sharkdp/bat/latest_releases_asset/bat-%2A-x86_64-unknown-linux-gnu.tar.gz \
	| tar -xz -C /usr/bin -f - --strip-components=1 --wildcards 'bat-*/bat'

#versioncommand: bat --version | cut -d' ' -f2
