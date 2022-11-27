#!/bin/bash

source common.sh


# recursively searches directories for a regex pattern while respecting your gitignore

curl -fsSL https://function61.com/app-dl/api/github.com/BurntSushi/ripgrep/latest_releases_asset/ripgrep-%2A-x86_64-unknown-linux-musl.tar.gz \
	| tar -C /usr/bin -xzf - --strip-components=1 --wildcards '*/rg'

#versioncommand: rg --version | cut -d ' ' -f2

