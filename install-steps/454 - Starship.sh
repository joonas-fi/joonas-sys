#!/bin/bash

source common.sh


# Shell prompt customization

curl -fsSL \
	https://function61.com/app-dl/api/github.com/starship/starship/latest_releases_asset/starship-x86_64-unknown-linux-gnu.tar.gz \
	| tar -C /usr/bin/ -xzf -

#versioncommand: starship --version | head -1 | cut -f 2 -d ' '
